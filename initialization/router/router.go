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
	router.DelTaskInit()
	router.UploadImgInit()
	router.ShopInit()
	router.BodyInit()
}
