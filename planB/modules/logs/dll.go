package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"planA/planB/initialization/golabl"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	LOG_LEVEL_DEBUG   = "DEBUG"
	LOG_LEVEL_INFO    = "INFO"
	LOG_LEVEL_WARNING = "WARNING"
	LOG_LEVEL_ERROR   = "ERROR"
	LOG_LEVEL_SUCCESS = "SUCCESS"
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

	dllPath := filepath.Join(golabl.Config.FileUrl.LogDll, "logger.dll")

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

// LoggingMiddleware 记录日志
func LoggingMiddleware(level string, str string) {
	initializeLoggerErr := InitializeLogger("logs")
	if initializeLoggerErr != nil {
		fmt.Println("初始化日志失败:", initializeLoggerErr)
		return
	}
	setLogTaskTypeErr := SetLogTaskType("task")
	if setLogTaskTypeErr != nil {
		fmt.Println("设置日志任务类型失败:", setLogTaskTypeErr)
		return
	}

	switch {
	case level == LOG_LEVEL_ERROR:
		fmt.Println(str)
		logErrorErr := LogError(str)
		if logErrorErr != nil {
			fmt.Println("记录错误日志失败:", logErrorErr)
			return
		}
	case level == LOG_LEVEL_WARNING:
		logWarningErr := LogWarning(str)
		if logWarningErr != nil {
			fmt.Println("记录警告日志失败:", logWarningErr)
			return
		}
	case level == LOG_LEVEL_SUCCESS:
		logSuccessErr := LogSuccess(str)
		if logSuccessErr != nil {
			fmt.Println("记录成功日志失败:", logSuccessErr)
			return
		}
	default:
		logInfoErr := LogInfo(str)
		if logInfoErr != nil {
			fmt.Println("记录信息日志失败:", logInfoErr)
			return
		}
	}
}
