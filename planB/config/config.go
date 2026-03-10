package config

import (
	"context"
	"encoding/json"
	"fmt"
	_configDll "planA/planB/modules/config"
	"planA/planB/tool"
	_type "planA/planB/type"
)

var (
	gDir string
)

// Init 初始化
// @param ctx context.Context 上下文
// @param dir string 配置文件目录
// @return _type.Config 配置文件信息
// @return error 错误信息
func Init(ctx context.Context, dir string) (_type.Config, error) {
	gDir = dir
	// 判断 ctx 是否取消
	checkContextErr := tool.CheckContext(ctx)
	// 判断 结果
	if checkContextErr != nil {
		// 返回 且 返回错误
		return _type.Config{}, checkContextErr
	}
	//读取配置文件
	var config _type.Config
	dll, initConfigDLLErr := _configDll.InitConfigDLL()
	if initConfigDLLErr != nil {
		return _type.Config{}, initConfigDLLErr
	}
	configJson, ReadConfigFileErr := dll.ReadConfigFile(dir, "config.yaml")
	if ReadConfigFileErr != nil {
		return _type.Config{}, fmt.Errorf("读取配置文件失败：%v", ReadConfigFileErr)
	}
	jsonUnmarshalErr := json.Unmarshal([]byte(configJson), &config)
	if jsonUnmarshalErr != nil {
		return _type.Config{}, fmt.Errorf("解析配置文件失败：%v", jsonUnmarshalErr)
	}
	return config, nil
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
// @return _type.DllFileUrl 文件路径配置
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
