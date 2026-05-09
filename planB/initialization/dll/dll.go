package dll

import (
	"planA/planB/initialization/dll/image"
	"planA/planB/initialization/dll/kfz"
	"planA/planB/initialization/dll/logs"
	"planA/planB/initialization/dll/pdd"
	"planA/planB/initialization/dll/xianYu"
)

// GetDllSetToG 获取DLL
func GetDllSetToG() error {
	// 初始化 PddDll
	getPddDllSetToGErr := pdd.GetPddDllSetToG()
	if getPddDllSetToGErr != nil {
		return getPddDllSetToGErr
	}
	// 初始化 ImageDll
	getImageDllSetToGErr := image.GetImageDllSetToG()
	if getImageDllSetToGErr != nil {
		return getImageDllSetToGErr
	}
	// 初始化 XianYuDll
	getXianYuDllSetToGErr := xianYu.GetXianYuDllSetToG()
	if getXianYuDllSetToGErr != nil {
		return getXianYuDllSetToGErr
	}
	// 初始化 LogrDll
	getLogrDllSetToGErr := logs.GetLogrDllSetToG()
	if getLogrDllSetToGErr != nil {
		return getLogrDllSetToGErr
	}
	// 获取KfzDll
	getKfzDllSetToGErr := kfz.GetKfzDllSetToG()
	if getKfzDllSetToGErr != nil {
		return getKfzDllSetToGErr
	}
	return nil
}
