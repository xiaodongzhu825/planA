package middle

import "planA/initialization/golabl"

// Init 初始化中间件
func Init() {
	golabl.Router.Use(Cors)
	golabl.Router.Use(Response)
}
