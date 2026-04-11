package logs

import (
	"planA/planB/initialization/golabl"
	"planA/planB/modules/logs"
)

// GetLogrDllSetToG 获取日志DLL
func GetLogrDllSetToG() error {
	dll, ensureLoggerDLLErr := logs.EnsureLoggerDLL(golabl.Config.FileUrl.LogDll)
	if ensureLoggerDLLErr != nil {
		return ensureLoggerDLLErr
	}
	golabl.LogDll = dll
	return nil
}
