package speed

import (
	"planA/planB/initialization/golabl"

	"golang.org/x/time/rate"
)

// Init 初始化 限速器
func Init() {
	//初始化限速器每秒18个
	golabl.Speed = rate.NewLimiter(rate.Limit(18), 1)
}
