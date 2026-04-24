package router

import (
	"planA/controller"
	"planA/initialization/golabl"
)

// DelTaskInit 删除任务初始化
func DelTaskInit() {
	delTaskRouter := golabl.Router.PathPrefix("/deltask").Subrouter()
	delTaskRouter.HandleFunc("/getDelTask", controller.GetDelTaskByPage).Methods("GET")                      // 分页查询删除任务
	delTaskRouter.HandleFunc("/getDelTaskByUserId/{id}", controller.GetDelTaskByPageByUserId).Methods("GET") // 分页查询删除任务-用户
	delTaskRouter.HandleFunc("/getDelTaskDetail/{id}", controller.GetDelTaskDetail).Methods("GET")           // 分页查询删除任务详情
}
