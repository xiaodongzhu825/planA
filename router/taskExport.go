package router

import (
	"planA/controller"
	"planA/initialization/golabl"
)

// TaskExportInit 任务导出初始化
func TaskExportInit() {
	taskExportRouter := golabl.Router.PathPrefix("/task/export").Subrouter()
	taskExportRouter.HandleFunc("/exportTaskDetail/{id}", controller.ExportTaskDetail).Methods("GET")                  // 导出任务详情（Excel/CSV）
	taskExportRouter.HandleFunc("/exportTaskDetail/{userId}/{id}", controller.ExportTaskDetailByUserId).Methods("GET") // 根据用户 ID导出任务详情（Excel/CSV）
	taskExportRouter.HandleFunc("/get", controller.GetExportTask).Methods("GET")                                       // 获取导出任务列表
	taskExportRouter.HandleFunc("/get/{userId}", controller.GetExportTaskUser).Methods("GET")                          // 根据用户 ID导出获取导出任务列表
}
