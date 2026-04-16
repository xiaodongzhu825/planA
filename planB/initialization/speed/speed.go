package speed

import (
	"planA/planB/initialization/golabl"

	"golang.org/x/time/rate"
)

// Init 初始化 限速器
func Init() {
	//默认为18
	speed := 18
	//根据平台设置速率
	switch golabl.Task.Header.ShopType {
	case "1":
		speed = golabl.Config.Speed.PddSpeed
	//case 2:
	case "5":
		speed = golabl.Config.Speed.XianyuSpeed
	default:
		speed = 18
	}
	//如果需要打水印，则速率下降为10
	if golabl.Task.Header.ShopMsg.WatermarkImgUrl != "" && golabl.Task.Header.ShopType == "1" {
		speed = golabl.Config.Speed.Watermark
		if speed == 0 {
			speed = 10
		}
	}
	//初始化限速器
	golabl.Speed = rate.NewLimiter(rate.Limit(speed), 1)
}
