package platform

import (
	"errors"
	pinDuoDuo "planA/planB/dispatcher/pinduoduo"
	"planA/planB/dispatcher/xianyu"
	"planA/planB/initialization/golabl"
)

// GetPlatformSetToG 获取平台并保存到全局变量中
func GetPlatformSetToG() error {
	switch golabl.Task.Header.ShopType {
	//case 2:
	//	return kongFuZi.NewKongfuzi(), nil
	case "1":
		golabl.Platform = pinDuoDuo.NewPinDuoDuo()
		return nil
	case "5":
		golabl.Platform = xianyu.NewXianYu()
		return nil
	default:
		return errors.New("错误！")
	}
}
