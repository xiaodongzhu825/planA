package speed

import (
	"fmt"
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
		fmt.Println("拼多多")
		speed = golabl.Config.Speed.PddSpeed
	//case 2:
	case "5":
		fmt.Println("闲鱼")
		speed = golabl.Config.Speed.XianyuSpeed
	default:
		speed = 18
	}
	//如果需要打水印，则速率下降为5
	if golabl.Task.Header.ShopMsg.WatermarkImgUrl != "" {
		speed = 5
	}
	//初始化限速器
	golabl.Speed = rate.NewLimiter(rate.Limit(speed), 1)
}
