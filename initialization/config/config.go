package config

import (
	"encoding/json"
	"fmt"
	"planA/initialization/golabl"
	_configDll "planA/modules/config"
	"planA/tool"
	_type "planA/type"
)

var (
	gDir string
)

// Init 初始化
// @param dir string 配置文件目录
// @return _type.Config 配置文件信息
// @return error 错误信息
func Init(dir string) error {
	gDir = dir
	// 判断 ctx 是否取消
	checkContextErr := tool.CheckContext(golabl.Ctx)
	// 判断 结果
	if checkContextErr != nil {
		// 返回 且 返回错误
		return checkContextErr
	}
	//读取配置文件
	var config _type.Config
	dll, initConfigDLLErr := _configDll.InitConfigDLL()
	if initConfigDLLErr != nil {
		return initConfigDLLErr
	}
	configJson, ReadConfigFileErr := dll.ReadConfigFile(dir, "config.yaml")
	if ReadConfigFileErr != nil {
		return fmt.Errorf("读取配置文件失败：%v", ReadConfigFileErr)
	}
	fmt.Println(configJson)
	jsonUnmarshalErr := json.Unmarshal([]byte(configJson), &config)
	if jsonUnmarshalErr != nil {
		return fmt.Errorf("解析配置文件失败：%v", jsonUnmarshalErr)
	}
	golabl.Config = config
	return nil
}

// GetPddClient 获取拼多多配置
// @return _type.PddConfig 拼多多配置
// @return error 错误信息
func GetPddClient() (_type.PddConfig, error) {
	//读取配置文件
	var config _type.Config
	dll, initConfigDLLErr := _configDll.InitConfigDLL()
	if initConfigDLLErr != nil {
		return _type.PddConfig{}, initConfigDLLErr
	}
	configJson, ReadConfigFileErr := dll.ReadConfigFile(gDir, "config.yaml")
	if ReadConfigFileErr != nil {
		return _type.PddConfig{}, fmt.Errorf("读取配置文件失败：%v", ReadConfigFileErr)
	}
	jsonUnmarshalErr := json.Unmarshal([]byte(configJson), &config)
	if jsonUnmarshalErr != nil {
		return _type.PddConfig{}, fmt.Errorf("解析配置文件失败：%v", jsonUnmarshalErr)
	}
	return config.PddConfig, nil
}

// GetFileUrlConfig 获取文件路径配置
// @return _type.FileUrl 文件路径配置
// @return error 错误信息
func GetFileUrlConfig() (_type.FileUrl, error) {
	//读取配置文件
	var config _type.Config
	dll, initConfigDLLErr := _configDll.InitConfigDLL()
	if initConfigDLLErr != nil {
		return _type.FileUrl{}, initConfigDLLErr
	}
	configJson, ReadConfigFileErr := dll.ReadConfigFile(gDir, "config.yaml")
	if ReadConfigFileErr != nil {
		return _type.FileUrl{}, fmt.Errorf("读取配置文件失败：%v", ReadConfigFileErr)
	}
	jsonUnmarshalErr := json.Unmarshal([]byte(configJson), &config)
	if jsonUnmarshalErr != nil {
		return _type.FileUrl{}, fmt.Errorf("解析配置文件失败：%v", jsonUnmarshalErr)
	}
	return config.FileUrl, nil
}

// GetAliveConfig 获取存活状态配置
// @return _type.Alive 存活状态配置
// @return error 错误信息
func GetAliveConfig() (_type.Alive, error) {
	//读取配置文件
	var config _type.Config
	dll, initConfigDLLErr := _configDll.InitConfigDLL()
	if initConfigDLLErr != nil {
		return _type.Alive{}, initConfigDLLErr
	}
	configJson, ReadConfigFileErr := dll.ReadConfigFile(gDir, "config.yaml")
	if ReadConfigFileErr != nil {
		return _type.Alive{}, fmt.Errorf("读取配置文件失败：%v", ReadConfigFileErr)
	}
	jsonUnmarshalErr := json.Unmarshal([]byte(configJson), &config)
	if jsonUnmarshalErr != nil {
		return _type.Alive{}, fmt.Errorf("解析配置文件失败：%v", jsonUnmarshalErr)
	}
	return config.Alive, nil
}
