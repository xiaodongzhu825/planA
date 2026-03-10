package router

import "planA/router"

// Init 初始化路由
func Init() {
	router.DefaultInit()
	router.TaskInit()
	router.TaskExportInit()
	router.StaticInit()
	router.AdmiinInir()
	router.Alive()
}
