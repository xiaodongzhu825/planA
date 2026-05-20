package router

import (
	"planA/planF/controller"
	"planA/planF/initialization/golabl"
)

// ApiInir 初始化api路由
func ApiInir() {
	adminExportRouter := golabl.Router.PathPrefix("/api").Subrouter()
	adminExportRouter.HandleFunc("/test", controller.Test).Methods("POST") // 删除 redis中指定任务
}
