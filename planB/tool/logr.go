package tool

import (
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/logs"
	"sync"
)

// LoggingMiddleware 记录日志
// 全局计数器，用于控制错误日志打印频率
var errorLogCounter = 0
var errorLogMutex sync.Mutex

func LoggingMiddleware(level string, str string) {
	m := golabl.LogDll
	initializeLoggerErr := logs.InitializeLogger(m, "logs")
	if initializeLoggerErr != nil {
		fmt.Println("初始化日志失败:", initializeLoggerErr)
		return
	}
	setLogTaskTypeErr := logs.SetLogTaskType(m, "task")
	if setLogTaskTypeErr != nil {
		fmt.Println("设置日志任务类型失败:", setLogTaskTypeErr)
		return
	}
	str = SteLog(str)

	// 设置打印间隔，例如每100条打印一次
	printInterval := 100 // 可以修改这个数字来调整打印频率

	switch {
	case level == logs.LOG_LEVEL_ERROR:
		// 控制打印频率
		shouldPrint := false
		errorLogMutex.Lock()
		errorLogCounter++
		if errorLogCounter%printInterval == 0 {
			shouldPrint = true
		}
		errorLogMutex.Unlock()

		if shouldPrint {
			fmt.Println(str)
		}

		logErrorErr := logs.LogError(m, str)
		if logErrorErr != nil {
			fmt.Println("记录错误日志失败:", logErrorErr)
			return
		}
	case level == logs.LOG_LEVEL_WARNING:
		logWarningErr := logs.LogWarning(m, str)
		if logWarningErr != nil {
			fmt.Println("记录警告日志失败:", logWarningErr)
			return
		}
	case level == logs.LOG_LEVEL_SUCCESS:
		logSuccessErr := logs.LogSuccess(m, str)
		if logSuccessErr != nil {
			fmt.Println("记录成功日志失败:", logSuccessErr)
			return
		}
	default:
		logInfoErr := logs.LogInfo(m, str)
		if logInfoErr != nil {
			fmt.Println("记录信息日志失败:", logInfoErr)
			return
		}
	}
}
