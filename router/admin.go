package router

import (
	"planA/controller"
	"planA/initialization/golabl"
)

// AdmiinInir 开发者管理
func AdmiinInir() {
	adminExportRouter := golabl.Router.PathPrefix("/admin").Subrouter()
	adminExportRouter.HandleFunc("/delRedisTask/{id}", controller.DelRedisTask).Methods("GET")   // 删除 redis中指定任务
	adminExportRouter.HandleFunc("/delMysqlTask/{id}", controller.DelMysqlTask).Methods("GET")   // 删除 mysql中指定任务
	adminExportRouter.HandleFunc("/delSqliteTask/{id}", controller.DelSqliteTask).Methods("GET") // 删除 sqlite中指定任务
	adminExportRouter.HandleFunc("/delSqliteTask/{id}", controller.DelSqliteTask).Methods("GET") // 删除 sqlite中指定任务
}
