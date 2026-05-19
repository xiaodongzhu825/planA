package dll

import (
	"planA/planE/initialization/dll/pdd"
)

// GetDllSetToG 获取DLL
func GetDllSetToG() error {
	// 初始化 PddDll
	getPddDllSetToGErr := pdd.GetPddDllSetToG()
	if getPddDllSetToGErr != nil {
		return getPddDllSetToGErr
	}
	return nil
}
