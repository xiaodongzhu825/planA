package router

import (
	"planA/controller"
	"planA/initialization/golabl"
)

// TaskInit 任务初始化
func TaskInit() {
	taskRouter := golabl.Router.PathPrefix("/task").Subrouter()
	taskRouter.HandleFunc("/create", controller.CreateTask).Methods("POST")            // 创建新任务
	taskRouter.HandleFunc("/pause/{id}", controller.PauseTask).Methods("GET")          // 暂停指定任务（任务ID）
	taskRouter.HandleFunc("/resume/{id}", controller.ResumeTask).Methods("GET")        // 恢复指定任务
	taskRouter.HandleFunc("/stop/{id}", controller.StopTask).Methods("GET")            // 停止指定任务
	taskRouter.HandleFunc("/over/{id}", controller.OverTask).Methods("GET")            // 完成任务
	taskRouter.HandleFunc("/get", controller.GetTask).Methods("GET")                   // 获取任务列表（支持查询参数）
	taskRouter.HandleFunc("/getByUserId", controller.GetTaskByUserId).Methods("GET")   // 根据用户 ID获取任务 获取任务列表（支持查询参数）
	taskRouter.HandleFunc("/setTaskBody", controller.SetTaskBody).Methods("POST")      // 设置任务执行内容
	taskRouter.HandleFunc("/del/{id}", controller.DelTask).Methods("GET")              // 删除任务
	taskRouter.HandleFunc("/b", controller.B).Methods("GET")                           // 运行B程序（特殊功能）
	taskRouter.HandleFunc("/header/get/{id}", controller.GetTaskHeader).Methods("GET") // 获取任务 header信息
}
