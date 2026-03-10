# logger.dll 使用教程
## 1. 创建DLL工具实例
### 加载DLL文件
```gotemplate
package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// LoggerDLL 封装 logger.dll 操作
type LoggerDLL struct {
	dll             *syscall.LazyDLL
	createLogger    *syscall.LazyProc
	createContext   *syscall.LazyProc
	logInfo         *syscall.LazyProc
	logError        *syscall.LazyProc
	logWarning      *syscall.LazyProc
	logSuccess      *syscall.LazyProc
	freeString      *syscall.LazyProc
	closeAllLoggers *syscall.LazyProc
}

// LoggerConfig logger配置结构
type LoggerConfig struct {
	LogDir          string `json:"log_dir"`
	SplitType       int    `json:"split_type"`
	RotateType      int    `json:"rotate_type"`
	MaxSize         int64  `json:"max_size"`
	MaxCount        int    `json:"max_count"`
	Level           int    `json:"level"`
	EnableCaller    bool   `json:"enable_caller"`
	DefaultTaskType string `json:"default_task_type"`
}

var loggerDLLInstance *LoggerDLL
var loggerHandle string
var loggerContextHandle string

// ensureLoggerDLL 确保logger DLL已加载
func ensureLoggerDLL() (*LoggerDLL, error) {
	if loggerDLLInstance != nil {
		return loggerDLLInstance, nil
	}

	// 检查是否在Windows平台
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("logger DLL only supported on Windows platform")
	}

	// logger.dll 位于 dll/logger.dll
	//dllPath := filepath.Join("modules", "logs", "logger.dll")
	dllPath := "D:\\www\\wwwroot\\planA\\modules\\logs\\logger.dll"

	// 检查文件是否存在
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		// 尝试从当前目录查找
		if _, err := os.Stat("logger.dll"); err == nil {
			dllPath = "logger.dll"
		} else {
			return nil, fmt.Errorf("logger DLL not found at %s", dllPath)
		}
	}

	dll := syscall.NewLazyDLL(dllPath)

	loggerDLLInstance = &LoggerDLL{
		dll:             dll,
		createLogger:    dll.NewProc("CreateLogger"),
		createContext:   dll.NewProc("CreateContextWithTaskType"),
		logInfo:         dll.NewProc("LogInfo"),
		logError:        dll.NewProc("LogError"),
		logWarning:      dll.NewProc("LogWarning"),
		logSuccess:      dll.NewProc("LogSuccess"),
		freeString:      dll.NewProc("FreeString"),
		closeAllLoggers: dll.NewProc("CloseAllLoggers"),
	}

	return loggerDLLInstance, nil
}

// cStr 将 C 字符串指针转换为 Go 字符串
func cStr(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	var b []byte
	for {
		c := *(*byte)(unsafe.Pointer(ptr))
		if c == 0 {
			break
		}
		b = append(b, c)
		ptr++
	}
	return string(b)
}

// InitializeLogger 初始化logger
func InitializeLogger(logDir string) error {
	m, err := ensureLoggerDLL()
	if err != nil {
		return err
	}

	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 创建logger配置
	config := LoggerConfig{
		LogDir:          logDir,
		SplitType:       1,                 // SplitByDay
		RotateType:      0,                 // RotateBySize
		MaxSize:         100 * 1024 * 1024, // 100MB
		MaxCount:        10,
		Level:           1, // LevelInfo - 只显示INFO及以上级别的日志
		EnableCaller:    true,
		DefaultTaskType: "main",
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 调用CreateLogger
	configPtr, _ := syscall.BytePtrFromString(string(configJSON))
	ret, _, _ := m.createLogger.Call(uintptr(unsafe.Pointer(configPtr)))

	if ret == 0 {
		return fmt.Errorf("创建logger失败")
	}

	// 获取logger句柄
	handle := cStr(ret)
	loggerHandle = handle

	// 释放返回的字符串
	m.freeString.Call(ret)

	// 创建默认上下文
	return createLoggerContext("main")
}

// createLoggerContext 创建带任务类型的logger上下文
func createLoggerContext(taskType string) error {
	m, err := ensureLoggerDLL()
	if err != nil {
		return err
	}

	if loggerHandle == "" {
		return fmt.Errorf("logger未初始化")
	}

	handlePtr, _ := syscall.BytePtrFromString(loggerHandle)
	taskTypePtr, _ := syscall.BytePtrFromString(taskType)

	ret, _, _ := m.createContext.Call(
		uintptr(unsafe.Pointer(handlePtr)),
		uintptr(unsafe.Pointer(taskTypePtr)),
	)

	if ret == 0 {
		return fmt.Errorf("创建logger上下文失败")
	}

	// 获取上下文句柄
	loggerContextHandle = cStr(ret)

	// 释放返回的字符串
	m.freeString.Call(ret)

	return nil
}

// SetLogTaskType 设置当前日志任务类型
func SetLogTaskType(taskType string) error {
	return createLoggerContext(taskType)
}

// LogInfo 记录信息日志
func LogInfo(message string) error {
	m, err := ensureLoggerDLL()
	if err != nil {
		return err
	}

	if loggerContextHandle == "" {
		return fmt.Errorf("logger上下文未初始化")
	}

	ctxPtr, _ := syscall.BytePtrFromString(loggerContextHandle)
	msgPtr, _ := syscall.BytePtrFromString(message)

	m.logInfo.Call(
		uintptr(unsafe.Pointer(ctxPtr)),
		uintptr(unsafe.Pointer(msgPtr)),
	)

	return nil
}

// LogError 记录错误日志
func LogError(message string) error {
	m, err := ensureLoggerDLL()
	if err != nil {
		return err
	}

	if loggerContextHandle == "" {
		return fmt.Errorf("logger上下文未初始化")
	}

	ctxPtr, _ := syscall.BytePtrFromString(loggerContextHandle)
	msgPtr, _ := syscall.BytePtrFromString(message)

	m.logError.Call(
		uintptr(unsafe.Pointer(ctxPtr)),
		uintptr(unsafe.Pointer(msgPtr)),
	)

	return nil
}

// LogWarning 记录警告日志
func LogWarning(message string) error {
	m, err := ensureLoggerDLL()
	if err != nil {
		return err
	}

	if loggerContextHandle == "" {
		return fmt.Errorf("logger上下文未初始化")
	}

	ctxPtr, _ := syscall.BytePtrFromString(loggerContextHandle)
	msgPtr, _ := syscall.BytePtrFromString(message)

	m.logWarning.Call(
		uintptr(unsafe.Pointer(ctxPtr)),
		uintptr(unsafe.Pointer(msgPtr)),
	)

	return nil
}

// LogSuccess 记录成功日志
func LogSuccess(message string) error {
	m, err := ensureLoggerDLL()
	if err != nil {
		return err
	}

	if loggerContextHandle == "" {
		return fmt.Errorf("logger上下文未初始化")
	}

	ctxPtr, _ := syscall.BytePtrFromString(loggerContextHandle)
	msgPtr, _ := syscall.BytePtrFromString(message)

	m.logSuccess.Call(
		uintptr(unsafe.Pointer(ctxPtr)),
		uintptr(unsafe.Pointer(msgPtr)),
	)

	return nil
}

// CloseLogger 关闭logger
func CloseLogger() error {
	m, err := ensureLoggerDLL()
	if err != nil {
		return err
	}

	ret, _, _ := m.closeAllLoggers.Call()
	if ret == 0 {
		return fmt.Errorf("关闭logger失败")
	}

	m.freeString.Call(ret)

	loggerHandle = ""
	loggerContextHandle = ""
	loggerDLLInstance = nil

	return nil
}

// GetLoggerHandle 获取当前logger句柄（用于外部调用）
func GetLoggerHandle() string {
	return loggerContextHandle
}

// IsLoggerInitialized 检查logger是否已初始化
func IsLoggerInitialized() bool {
	return loggerHandle != "" && loggerContextHandle != ""
}

// SetConsoleOutput 设置控制台输出开关
func SetConsoleOutput(enabled bool) {
	if enabled {
		os.Setenv("LOG_CONSOLE", "true")
	} else {
		os.Setenv("LOG_CONSOLE", "false")
	}
}

// LogWithLevel 带级别的日志记录，可以精确控制显示
func LogWithLevel(level, message string, showConsole bool) {
	if !IsLoggerInitialized() {
		return
	}

	switch level {
	case "ERROR":
		LogError(message)
	case "WARNING":
		LogWarning(message)
	case "SUCCESS":
		LogSuccess(message)
	case "INFO":
		LogInfo(message)
	default:
		LogInfo(message)
	}
}

// LogOnlyFile 仅写入文件，不输出到控制台
func LogOnlyFile(level, message string) {
	// 临时禁用控制台输出
	os.Setenv("LOG_CONSOLE", "false")
	LogWithLevel(level, message, false)
}

// LogConsoleAndFile 同时输出到控制台和文件
func LogConsoleAndFile(level, message string) {
	// 临时启用控制台输出
	os.Setenv("LOG_CONSOLE", "true")
	LogWithLevel(level, message, true)
}

```

# 接口详情
## 创建日志器--CreateLogger
### 请求信息
```gotemplate
dll.CreateLogger(configJSON)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明           |
|--|--|--|--------------|
| configJSON | string | 是 | 配置信息JSON字符串  |
#### 配置JSON结构
```json
{
  "log_dir": "/path/to/logs",
  "split_type": 0,
  "rotate_type": 0,
  "max_size": 104857600,
  "max_count": 30,
  "level": 1,
  "enable_caller": true,
  "default_task_type": "main"
}
```
#### 参数说明：
```text
log_dir: 日志目录路径
split_type: 分片方式（0=按月，1=按天，2=按小时，3=按分钟，4=按秒）
rotate_type: 轮转方式（0=按大小，1=按数量）
max_size: 最大文件大小（字节），仅在rotate_type=0时有效
max_count: 最大文件数量，仅在rotate_type=1时有效
level: 日志级别（0=SUCCESS，1=INFO，2=WARNING，3=ERROR）
enable_caller: 是否启用调用者信息
default_task_type: 默认任务类型
```
### 响应示例
```json
"错误: 创建日志目录失败: permission denied"
```

## 创建带任务类型的上下文--CreateContextWithTaskType
### 请求信息
```gotemplate
dll.CreateContextWithTaskType(loggerHandle, taskType)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明           |
|--|--|--|--------------|
| loggerHandle | string | 是 | 日志器句柄  |
| taskType | string | 是 | 任务类型  |
### 响应示例
```json
"ctx_1645497600000000000"
```
#### 错误响应示例
```json
"错误: 无效的logger句柄"
```

## 记录信息日志--LogInfo
### 请求信息
```gotemplate
dll.LogInfo(ctxHandle, message)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明           |
|--|--|--|--------------|
| ctxHandle | string | 是 | 上下文句柄  |
| message | string | 是 | 日志消息  |
### 响应示例
```text
无返回值，日志将写入到对应的日志文件中。
```

## 记录错误日志--LogError
### 请求信息
```gotemplate
dll.LogError(ctxHandle, message)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明           |
|--|--|--|--------------|
| ctxHandle | string | 是 | 上下文句柄  |
| message | string | 是 | 日志消息  |
### 响应示例
```text
无返回值，日志将写入到对应的日志文件中。
```

## 记录警告日志--LogWarning
### 请求信息
```gotemplate
dll.LogWarning(ctxHandle, message)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明           |
|--|--|--|--------------|
| ctxHandle | string | 是 | 上下文句柄  |
| message | string | 是 | 日志消息  |
### 响应示例
```text
无返回值，日志将写入到对应的日志文件中。
```

## 记录成功日志--LogSuccess
### 请求信息
```gotemplate
dll.LogSuccess(ctxHandle, message)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明           |
|--|--|--|--------------|
| ctxHandle | string | 是 | 上下文句柄  |
| message | string | 是 | 日志消息  |
### 响应示例
```text
无返回值，日志将写入到对应的日志文件中。
```

## 获取日志条目--GetLogs
### 请求信息
```gotemplate
dll.GetLogs(loggerHandle, configJSON)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明        |
|--|--|--|-----------|
| loggerHandle | string | 是 | 日志器句柄     |
| configJSON | string | 是 | 查询配置JSON  |
#### 查询配置JSON结构
```json
{
  "level": 1,
  "task_type": "main",
  "start_time": "2024-01-01 00:00:00",
  "end_time": "2024-01-31 23:59:59",
  "max_entries": 1000
}
```
#### 参数说明：
```text
level: 日志级别（-1表示所有级别）
task_type: 任务类型（空字符串表示所有任务类型）
start_time: 开始时间（格式: 2006-01-02 15:04:05）
end_time: 结束时间（格式: 2006-01-02 15:04:05）
max_entries: 最大返回条目数（0表示使用默认值1000）
```
### 响应示例
```json
{
  "count": 125,
  "entries": [
    {
      "timestamp": "2024-01-15 10:30:45.123",
      "level": "INFO",
      "task_type": "main",
      "caller": "logger.go:256",
      "message": "系统启动完成"
    },
    {
      "timestamp": "2024-01-15 10:31:15.456",
      "level": "ERROR",
      "task_type": "backup",
      "caller": "backup.go:89",
      "message": "备份文件失败: 磁盘空间不足"
    }
  ]
}
```
### 错误响应示例
```json
{"error": "无效的logger句柄"}
```

## 获取日志文件列表--GetLogFiles
### 请求信息
```gotemplate
dll.GetLogFiles(loggerHandle)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明        |
|--|--|--|-----------|
| loggerHandle | string | 是 | 日志器句柄     |
### 响应示例
```json
{
  "count": 8,
  "files": [
    {
      "level": "INFO",
      "task_type": "main",
      "file_name": "INFO-main-2024-01.logs",
      "file_size": 1048576,
      "mod_time": "2024-01-15 10:30:45"
    },
    {
      "level": "ERROR",
      "task_type": "backup",
      "file_name": "ERROR-backup-2024-01.logs",
      "file_size": 51200,
      "mod_time": "2024-01-15 10:31:15"
    }
  ]
}
```
### 错误响应示例
```json
{"error": "无效的logger句柄"}
```

## 获取版本信息--GetVersion
### 请求信息
```gotemplate
dll.GetVersion()
```
### 请求参数
```text
无参数
```
### 响应示例
```json
"v1"
```

## 关闭所有日志器--CloseAllLoggers
### 请求信息
```gotemplate
dll.CloseAllLoggers()
```
### 请求参数
```text
无参数
```
### 响应示例
```json
"成功关闭所有logger"
```
### 错误响应示例
```json
"关闭了5个logger，其中1个出错，最后错误: close file error"
```

## 释放C字符串内存--FreeString
### 请求信息
```gotemplate
dll.FreeString(str)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明        |
|--|--|--|-----------|
| str | string | 是 | 需要释放的字符串  |