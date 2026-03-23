package config

import (
	"encoding/json"
	"fmt"
	"planA/planB/initialization/golabl"
	planBConfig "planA/planB/modules/config"
	"planA/tool"
	planAtype "planA/type"
)

// GetConfigSetToG 获取配置文件并保存到全局变量中
// @return error 错误信息
func GetConfigSetToG() error {
	// 检查全局 CTX 是否失效 以防止重复初始化和 ctx 失效 导致程序崩溃
	checkContextErr := tool.CheckContext(golabl.Ctx)
	if checkContextErr != nil {
		// 返回 且 返回错误
		return checkContextErr
	}

	//读取配置文件
	var config planAtype.Config

	// 加载 config.dll
	dll, initConfigDLLErr := planBConfig.InitConfigDLL()
	if initConfigDLLErr != nil {
		return initConfigDLLErr
	}

	// 读取配置文件
	configJson, ReadConfigFileErr := dll.ReadConfigFile("", "config.yaml")
	if ReadConfigFileErr != nil {
		return fmt.Errorf("读取配置文件失败：%v", ReadConfigFileErr)
	}
	// 转换配置文件到 JSON
	jsonUnmarshalErr := json.Unmarshal([]byte(configJson), &config)
	if jsonUnmarshalErr != nil {
		return fmt.Errorf("解析配置文件失败：%v", jsonUnmarshalErr)
	}

	// 保存到全局变量
	golabl.Config = config
	// 返回
	return nil
}
