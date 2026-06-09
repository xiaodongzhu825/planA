package router

import (
	"planA/controller"
	"planA/initialization/golabl"
)

// BodyInit 任务体信息
func BodyInit() {
	taskRouter := golabl.Router.PathPrefix("/body").Subrouter()
	taskRouter.HandleFunc("/getOneBody/{taskId}", controller.GetTbOneBodyWait).Methods("GET") // 获取body信息
	taskRouter.HandleFunc("/insterOneBodyOver", controller.InsertTbBodyOver).Methods("POST")  // 插入bodyOver
}
