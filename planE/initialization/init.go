package initialization

import (
	"context"
	"fmt"
	"planA/planE/initialization/config"
	"planA/planE/initialization/dll"
	"planA/planE/initialization/golabl"
)

// Init 初始化
func Init() error {
	//初始化上下文
	if golabl.Ctx == nil {
		golabl.Ctx = context.Background()
	}
	// 初始化配置文件
	if configErr := config.GetConfigSetToG(); configErr != nil {
		return fmt.Errorf("初始化配置文件失败：%v", configErr)
	}

	// 初始化 DLL
	if dllErr := dll.GetDllSetToG(); dllErr != nil {
		return fmt.Errorf("初始化DLL失败: %v", dllErr)
	}

	return nil
}
