package dll

import (
	"planA/planD/initialization/dll/kfz"
	"planA/planD/initialization/dll/pdd"
)

// GetDllSetToG 获取DLL
func GetDllSetToG() error {
	// 初始化 PddDll
	getPddDllSetToGErr := pdd.GetPddDllSetToG()
	if getPddDllSetToGErr != nil {
		return getPddDllSetToGErr
	}

	// 初始化 孔夫子DLL
	getKfzDllSetToGErr := kfz.GetKfzDllSetToG()
	if getKfzDllSetToGErr != nil {
		return getKfzDllSetToGErr
	}
	return nil
}
