package router

import (
	"planA/controller"
	"planA/initialization/golabl"
)

// Alive 初始化工具
func Alive() {
	aliveRouter := golabl.Router.PathPrefix("/alive").Subrouter()
	aliveRouter.HandleFunc("/get", controller.GetServiceAliveList).Methods("GET") // 获取服务存活状态列表
}
