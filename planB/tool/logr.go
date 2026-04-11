package tool

import (
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/logs"
)

// LoggingMiddleware 记录日志
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
	switch {
	case level == logs.LOG_LEVEL_ERROR:
		fmt.Println(str)
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
