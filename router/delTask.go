package router

import (
	"planA/controller"
	"planA/initialization/golabl"
)

// DelTaskInit 删除任务初始化
func DelTaskInit() {
	delTaskRouter := golabl.Router.PathPrefix("/deltask").Subrouter()
	delTaskRouter.HandleFunc("/getDelTask", controller.GetDelTaskByPage).Methods("GET")                                // 分页查询删除任务
	delTaskRouter.HandleFunc("/getDelTaskByUserId/{id}", controller.GetDelTaskByPageByUserId).Methods("GET")           // 分页查询删除任务-用户
	delTaskRouter.HandleFunc("/getDelTaskDetail/{id}", controller.GetDelTaskDetail).Methods("GET")                     // 分页查询删除任务详情
	delTaskRouter.HandleFunc("/createTbDelTask", controller.CreateTbDelTask).Methods("POST")                           // 创建淘宝删除任务
	delTaskRouter.HandleFunc("/createTbDelTaskDetails", controller.CreateTbDelTaskDetails).Methods("POST")             // 插入淘宝删除任务
	delTaskRouter.HandleFunc("/updateTbDelTaskDetailsStatus", controller.UpdateTbDelTaskDetailsStatus).Methods("POST") // 修改指定淘宝删除任务详情状态
	delTaskRouter.HandleFunc("/updateTbDelTaskProgress", controller.UpdateTbDelTaskProgress).Methods("POST")           // 修改指定淘宝任务进度
	delTaskRouter.HandleFunc("/updateTbDelTaskStatus", controller.UpdateTbDelTaskStatus).Methods("POST")               // 修改指定淘宝任务状态
	delTaskRouter.HandleFunc("/getTbDelTaskDetailsWait/{id}", controller.GetTbDelTaskDetailsWait).Methods("GET")       // 获取淘宝删除任务详情-待处理
	delTaskRouter.HandleFunc("/getTbDelTaskByTaskId/{id}", controller.GetTbDelTaskByTaskId).Methods("GET")             // 获取任务数据
}
