package kfz

import (
	"planA/planB/initialization/golabl"
	"planA/planB/modules/kfz"
)

// GetKfzDllSetToG 获取孔夫子DLL
func GetKfzDllSetToG() error {
	// 初始化 KfzDll
	kfzDll, err := kfz.InitKfzDll(golabl.Config.FileUrl.KfzDll)
	if err != nil {
		return err
	}
	golabl.KfzDll = kfzDll
	return nil
}
