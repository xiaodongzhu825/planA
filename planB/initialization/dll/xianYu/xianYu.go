package xianYu

import (
	"planA/planB/initialization/golabl"
	"planA/planB/modules/xianYu"
)

// GetXianYuDllSetToG 获取闲鱼DLL
func GetXianYuDllSetToG() error {
	xianYuDll, xianYuDllErr := xianYu.InitXianYuDll(golabl.Config.FileUrl.XianYuDll)
	if xianYuDllErr != nil {
		return xianYuDllErr
	}
	golabl.XianYuDll = xianYuDll
	return nil
}
