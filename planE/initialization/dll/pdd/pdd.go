package pdd

import (
	"planA/planE/initialization/golabl"
	"planA/planE/modules/pdd"
)

// GetPddDllSetToG 获取拼多多DLL
func GetPddDllSetToG() error {
	// 初始化 PddDll
	pddDll, err := pdd.InitPddDll(golabl.Config.FileUrl.PddDll)
	if err != nil {
		return err
	}
	golabl.PddDll = pddDll
	return nil
}
