package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"planA/controlState/lock"
	"planA/initialization/config"
	"planA/modules/logs"
	"planA/service"
	_type "planA/type"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	_redis "github.com/go-redis/redis/v8"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	modntdll    = syscall.NewLazyDLL("ntdll.dll")

	procOpenProcess      = modkernel32.NewProc("OpenProcess")
	procCloseHandle      = modkernel32.NewProc("CloseHandle")
	procNtSuspendProcess = modntdll.NewProc("NtSuspendProcess")
	procNtResumeProcess  = modntdll.NewProc("NtResumeProcess")
)

const (
	PROCESS_SUSPEND_RESUME = 0x0800
)

// RunTaskWorker 启动 B程序
// @param taskId 任务ID
// @return string 进程ID
// @return error 错误
func RunTaskWorker(taskId string) (string, error) {
	// 1. 尝试加锁
	if !lock.TryLock(taskId) {
		return "", fmt.Errorf("taskId %s 已被上锁，跳过B程序执行", taskId)
	}
	// 2 加锁成功：执行B程序，确保defer释放锁（即使执行出错也能解锁）
	defer lock.DestroyLock(taskId)

	// 3 判断pid是否存在
	processId, getProcessIdErr := service.GetProcessId(taskId)
	// 检查是否有错误（排除redis key不存在的情况）
	if getProcessIdErr != nil && !errors.Is(getProcessIdErr, _redis.Nil) {
		return "", getProcessIdErr
	}

	// 4 判断pid是否存在
	if processId != "" {
		//验证进程是否真实存在
		if !IsProcessExistWindows(processId) {
			// 删除 header中的pid
			deleteProcessIdErr := service.DeleteProcessId(taskId)
			if deleteProcessIdErr != nil {
				return "", deleteProcessIdErr
			}
		} else {
			// 返回 pid
			return processId, nil
		}
	}

	// 5 判断任务状态
	taskStatus, getTaskStatusErr := service.GetTaskHeader(taskId)
	if getTaskStatusErr != nil {
		return "", getTaskStatusErr
	}
	if taskStatus.Status == _type.TaskStatusPaused {
		return "", fmt.Errorf("任务暂停中，请先尝试恢复")
	}
	if taskStatus.Status == _type.TaskStatusStopped || taskStatus.Status == _type.TaskStatusOver {
		return "", fmt.Errorf("任务已结束,不支持重启")
	}

	// 6 启动B程序
	_, CallSendPublishingERR := CallSendPublishing(taskId)
	if CallSendPublishingERR != nil {
		return "", CallSendPublishingERR
	}
	return processId, nil
}

// SuspendProcess 暂停指定PID的进程
// @param taskId 任务ID
// @return error 错误
func SuspendProcess(taskId string) error {
	// 1. 查询PID
	processId, getProcessIdErr := service.GetProcessId(taskId)
	if getProcessIdErr != nil {
		return getProcessIdErr
	}

	// 2. 将字符串转换为整数
	pid, err := strconv.Atoi(processId)
	if err != nil {
		return fmt.Errorf("PID格式错误: %s, 错误: %v", processId, err)
	}

	// 3. 检查PID是否有效
	if pid <= 0 {
		return fmt.Errorf("PID必须为正整数")
	}

	// 4. 打开进程
	hProcess, _, err := procOpenProcess.Call(
		PROCESS_SUSPEND_RESUME,
		uintptr(0),
		uintptr(pid),
	)
	if hProcess == 0 {
		return fmt.Errorf("打开进程失败: %v", err)
	}
	defer procCloseHandle.Call(hProcess)

	// 5. 暂停进程
	callSstatus, _, _ := procNtSuspendProcess.Call(hProcess)
	if callSstatus != 0 {
		return fmt.Errorf("NtSuspendProcess 失败: 0x%X", callSstatus)
	}

	// 6. 修改Header中的状态
	status := int64(_type.TaskStatusPaused)
	updateHeaderStatusErr := service.UpdateHeaderStatus(taskId, status)
	if updateHeaderStatusErr != nil {
		return updateHeaderStatusErr
	}
	return nil
}

// ResumeProcess 恢复指定PID的进程
// @param taskId 任务ID
// @return error 错误
func ResumeProcess(taskId string) error {
	// 1. 查询PID
	processId, getProcessIdErr := service.GetProcessId(taskId)
	if getProcessIdErr != nil {
		return getProcessIdErr
	}

	// 2. 将字符串转换为整数
	pid, err := strconv.Atoi(processId)
	if err != nil {
		return fmt.Errorf("PID格式错误: %s, 错误: %v", processId, err)
	}

	// 3. 检查PID是否有效
	if pid <= 0 {
		return fmt.Errorf("PID必须为正整数")
	}

	// 4. 打开进程
	hProcess, _, err := procOpenProcess.Call(
		PROCESS_SUSPEND_RESUME,
		uintptr(0),
		uintptr(pid),
	)
	if hProcess == 0 {
		return fmt.Errorf("打开进程失败: %v", err)
	}
	defer procCloseHandle.Call(hProcess)

	// 5. 恢复进程
	callStatus, _, _ := procNtResumeProcess.Call(hProcess)
	if callStatus != 0 {
		return fmt.Errorf("NtResumeProcess 失败: 0x%X", callStatus)
	}

	// 6. 修改Header中的状态
	status := int64(_type.TaskStatusRunning)
	updateHeaderStatusErr := service.UpdateHeaderStatus(taskId, status)
	if updateHeaderStatusErr != nil {
		return updateHeaderStatusErr
	}

	return nil
}

// StopTask 停止任务
// @param taskId 队列名称
// @return error 错误
func StopTask(taskId string) error {
	// 1. 恢复B程序避免程序处于暂停状态
	resumeProcessErr := ResumeProcess(taskId)
	if resumeProcessErr != nil {
		return resumeProcessErr
	}

	// 2. 修改 Header中的状态 并且 删除 bodyWait 中的数据
	stopTaskErr := service.StopTask(taskId)
	if stopTaskErr != nil {
		return stopTaskErr
	}
	return nil
}
func CallSendPublishing(taskId string) (*os.Process, error) {
	// 1. 基础入参校验
	if taskId == "" {
		return nil, errors.New("队列名称qName不能为空")
	}

	// 先在Redis中创建一个占位符，表示进程即将启动
	placeholderPID := "starting"
	setProcessIdErr := service.SetProcessId(taskId, placeholderPID)
	if setProcessIdErr != nil {
		errMsg := fmt.Sprintf("保存进程占位符到Redis失败: %v, taskId: %s", setProcessIdErr, taskId)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 2. 构建并验证程序路径
	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		errMsg := fmt.Sprintf("获取文件路径配置失败: %v", getFileUrlConfigErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	programPath := fileUrlConfig.BFileName

	// 3. 验证程序路径是否存在
	absProgramPath, err := filepath.Abs(programPath)
	if err != nil {
		errMsg := fmt.Sprintf("转换程序路径为绝对路径失败: %s, 原始路径: %s", err, programPath)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 4. 验证程序路径
	_, statErr := os.Stat(absProgramPath)
	if statErr != nil {
		errMsg := fmt.Sprintf("程序文件不存在或无访问权限: %s, 错误: %v", absProgramPath, statErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 5. 修复后的PowerShell脚本
	psScript := fmt.Sprintf(`
# 设置错误捕获模式
$ErrorActionPreference = "Stop"
$programPath = "%s"
$arguments = "%s"

try {
    # 再次验证程序存在性
    if (-not (Test-Path $programPath -PathType Leaf)) {
        throw "程序文件不存在: $programPath"
    }
    
    # 构建进程启动信息
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $programPath
    $psi.Arguments = $arguments
    $psi.UseShellExecute = $true
    $psi.WindowStyle = 'Normal'
    $psi.WorkingDirectory = (Split-Path $programPath -Parent)  # 设置工作目录为程序所在目录
    
    # 启动进程
    Write-Host "开始启动程序: $programPath 参数: $arguments"
    $process = [System.Diagnostics.Process]::Start($psi)
    Write-Host "程序启动成功，PID: $($process.Id)"
    
    # 等待窗口出现，设置超时时间
    $timeout = 3000
    $startTime = Get-Date
    $hwnd = $null
    
    # 等待窗口句柄不为空
    while ($true) {
        try {
            $process.Refresh()
            $hwnd = $process.MainWindowHandle
            if ($hwnd -and $hwnd -ne [IntPtr]::Zero) {
                break
            }
        } catch {
            # 忽略刷新错误
        }
        
        if (((Get-Date) - $startTime).TotalMilliseconds -ge $timeout) {
            Write-Warning "等待窗口句柄超时 (PID: $($process.Id))"
            break
        }
        Start-Sleep -Milliseconds 50
    }
    
    # 尝试将窗口前置（仅在成功获取窗口句柄时）
    if ($hwnd -and $hwnd -ne [IntPtr]::Zero) {
        try {
            Add-Type @"
                using System;
                using System.Runtime.InteropServices;
                public class WindowHelper {
                    [DllImport("user32.dll")]
                    public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
                    [DllImport("user32.dll")]
                    public static extern bool SetForegroundWindow(IntPtr hWnd);
                    [DllImport("user32.dll")]
                    public static extern bool AllowSetForegroundWindow(uint dwProcessId);
                }
"@
            
            [WindowHelper]::AllowSetForegroundWindow($process.Id)
            [WindowHelper]::ShowWindow($hwnd, 9)  # 9 = SW_RESTORE
            [WindowHelper]::SetForegroundWindow($hwnd)
            Write-Host "成功将窗口前置，句柄: $hwnd"
        } catch {
            Write-Warning "窗口前置操作失败: $_"
            # 不抛出异常，继续执行
        }
    } else {
        Write-Warning "未能获取窗口句柄，程序可能没有窗口界面 (PID: $($process.Id))"
    }
    
    # 输出PID供Go解析
    Write-Output $process.Id
} catch {
    # 捕获所有异常并输出
    Write-Error "启动程序失败: $_"
    exit 1  # 返回非0退出码
}
`, absProgramPath, taskId)

	// 6. 执行PowerShell命令，同时捕获stdout和stderr
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x00000010, // CREATE_NEW_CONSOLE - 创建新控制台
		}
	}

	// 7. 同时捕获标准输出和标准错误
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 8. 执行命令
	runErr := cmd.Run()

	// 9. 输出所有PowerShell的输出
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// 记录标准输出（调试用）
	if stdoutStr != "" {
		logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, fmt.Sprintf("PowerShell标准输出: %s", stdoutStr))
	}

	// 记录标准错误（警告信息不算错误）
	if stderrStr != "" {
		// 检查是否为警告信息
		if strings.Contains(stderrStr, "WARNING:") {
			logs.LoggingMiddleware(logs.LOG_LEVEL_WARNING, fmt.Sprintf("PowerShell警告: %s", stderrStr))
		} else {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("PowerShell错误: %s", stderrStr))
		}
	}

	// 10. 检查命令执行是否失败
	if runErr != nil {
		errMsg := fmt.Sprintf("PowerShell执行失败: %v, 错误输出: %s", runErr, stderrStr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 11. 解析PID（增加校验）
	var pid uint32
	lines := strings.Split(stdoutStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 跳过非纯数字行（如 Write-Host 的输出）
		pidInt, err := strconv.Atoi(line)
		if err == nil && pidInt > 0 {
			pid = uint32(pidInt)
			logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, fmt.Sprintf("成功解析PID: %d", pid))
			break // 找到有效PID立即退出
		}
	}

	if pid == 0 {
		errMsg := fmt.Sprintf("未解析到有效PID，PowerShell输出: %s, 错误输出: %s", stdoutStr, stderrStr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 12. 更新Redis中的PID
	processID := fmt.Sprintf("%d", pid)
	setProcessIdErr = service.SetProcessId(taskId, processID)
	if setProcessIdErr != nil {
		errMsg := fmt.Sprintf("更新进程PID到Redis失败: %v, taskId: %s, PID: %d", setProcessIdErr, taskId, pid)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 13. 返回进程句柄
	process := &os.Process{Pid: int(pid)}
	return process, nil
}

// CallSendPublishing1 启动程序
func CallSendPublishing1(taskId string) (*os.Process, error) {
	// 1. 基础入参校验
	if taskId == "" {
		return nil, errors.New("队列名称qName不能为空")
	}

	// 先在Redis中创建一个占位符，表示进程即将启动
	placeholderPID := "starting"
	setProcessIdErr := service.SetProcessId(taskId, placeholderPID)
	if setProcessIdErr != nil {
		errMsg := fmt.Sprintf("保存进程占位符到Redis失败: %v, taskId: %s", setProcessIdErr, taskId)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 2. 构建并验证程序路径
	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		errMsg := fmt.Sprintf("获取文件路径配置失败: %v", getFileUrlConfigErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	programPath := fileUrlConfig.BFileName

	// 3. 验证程序路径是否存在
	absProgramPath, err := filepath.Abs(programPath)
	if err != nil {
		errMsg := fmt.Sprintf("转换程序路径为绝对路径失败: %s, 原始路径: %s", err, programPath)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	// 4. 验证程序路径
	_, statErr := os.Stat(absProgramPath)
	if statErr != nil {
		errMsg := fmt.Sprintf("程序文件不存在或无访问权限: %s, 错误: %v", absProgramPath, statErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 5. 重构PowerShell脚本，增加错误捕获和详细日志
	psScript := fmt.Sprintf(`
        # 设置错误捕获模式
        $ErrorActionPreference = "Stop"
        $programPath = "%s"
        $arguments = "%s"
        
        try {
            # 再次验证程序存在性
            if (-not (Test-Path $programPath -PathType Leaf)) {
                throw "程序文件不存在: $programPath"
            }
            
            # 构建进程启动信息
            $psi = New-Object System.Diagnostics.ProcessStartInfo
            $psi.FileName = $programPath
            $psi.Arguments = $arguments
            $psi.UseShellExecute = $true
            $psi.WindowStyle = 'Normal'
            $psi.WorkingDirectory = (Split-Path $programPath -Parent)  # 设置工作目录为程序所在目录
            
            # 启动进程
            Write-Host "开始启动程序: $programPath 参数: $arguments"
            $process = [System.Diagnostics.Process]::Start($psi)
            Write-Host "程序启动成功，PID: $($process.Id)"
            
            # 等待窗口句柄（非必须，但保留原有逻辑）
            $timeout = 3000
            $startTime = Get-Date
            while ($process.MainWindowHandle -eq [IntPtr]::Zero -and ((Get-Date) - $startTime).TotalMilliseconds -lt $timeout) {
                Start-Sleep -Milliseconds 50
                $process.Refresh()
            }
            
            # 尝试将窗口前置
            if ($process.MainWindowHandle -ne [IntPtr]::Zero) {
                Add-Type @"
                    using System;
                    using System.Runtime.InteropServices;
                    public class WindowHelper {
                        [DllImport("user32.dll")]
                        public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
                        [DllImport("user32.dll")]
                        public static extern bool SetForegroundWindow(IntPtr hWnd);
                        [DllImport("user32.dll")]
                        public static extern bool AllowSetForegroundWindow(uint dwProcessId);
                    }
"@
                [WindowHelper]::AllowSetForegroundWindow($process.Id)
                [WindowHelper]::ShowWindow($process.MainWindowHandle, 9)
                [WindowHelper]::SetForegroundWindow($process.MainWindowHandle)
                Write-Host "程序窗口已前置，句柄: $($process.MainWindowHandle)"
            } else {
                Write-Warning "未能获取窗口句柄，但进程已启动 (PID: $($process.Id))"
            }
            
            # 输出PID供Go解析
            Write-Output $process.Id
        } catch {
            # 捕获所有异常并输出
            Write-Error "启动程序失败: $_"
            exit 1  # 返回非0退出码
        }
    `, absProgramPath, taskId)

	// 6. 执行PowerShell命令，同时捕获stdout和stderr
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x00000010, // CREATE_NEW_CONSOLE - 创建新控制台
		}
	}

	// 7. 同时捕获标准输出和标准错误
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 8. 执行命令
	runErr := cmd.Run()

	// 9. 输出所有PowerShell的输出
	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	if stderrStr != "" {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("PowerShell标准错误: %s", stderrStr))
	}

	// 10. 检查命令执行是否失败
	if runErr != nil {
		errMsg := fmt.Sprintf("PowerShell执行失败: %v, 标准错误: %s", runErr, stderrStr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 11. 解析PID（增加校验）
	var pid uint32
	lines := strings.Split(stdoutStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pidInt, err := strconv.Atoi(line)
		if err == nil && pidInt > 0 {
			pid = uint32(pidInt)
			break // 找到有效PID立即退出
		}
	}
	if pid == 0 {
		errMsg := fmt.Sprintf("未解析到有效PID，PowerShell输出: %s", stdoutStr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 6. 更新Redis中的PID
	processID := fmt.Sprintf("%d", pid)
	setProcessIdErr = service.SetProcessId(taskId, processID)
	if setProcessIdErr != nil {
		errMsg := fmt.Sprintf("更新进程PID到Redis失败: %v, taskId: %s, PID: %d", setProcessIdErr, taskId, pid)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 8. 返回进程句柄
	process := &os.Process{Pid: int(pid)}
	return process, nil
}

// IsProcessExistWindows 检查Windows进程是否存在
func IsProcessExistWindows(pid string) bool {
	if pid == "" {
		return false
	}
	pid64, err := strconv.ParseInt(pid, 10, 0)
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_WARNING, fmt.Sprintf("转换进程PID:%v失败: %v", pid, err))
		return false
	}
	pidInt := int(pid64)
	// 使用 tasklist命令检查进程是否存在
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pidInt))
	output, err := cmd.Output()
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_WARNING, fmt.Sprintf("检查进程PID:%v失败: %v", pid, err))
		return false
	}
	return strings.Contains(strings.ToLower(string(output)), fmt.Sprintf("%d", pidInt))
}
