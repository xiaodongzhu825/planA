package controller

import (
	"net/http"
	"planA/tool"

	"github.com/gorilla/mux"
)

// DelRedisTask 删除redis中指定任务
func DelRedisTask(httpMsg http.ResponseWriter, data *http.Request) {
	// 从路径参数获取 id
	vars := mux.Vars(data)
	taskId := vars["id"]
	// 验证 taskId
	if taskId == "" {
		errMsg := "任务 ID不能为空"
		tool.Error(httpMsg, errMsg, http.StatusBadRequest)
		return
	}
	// 删除任务

	tool.Session(httpMsg, taskId)
}

// DelMysqlTask 删除mysql中指定任务
func DelMysqlTask(httpMsg http.ResponseWriter, data *http.Request) {
	tool.Session(httpMsg, "删除mysql中指定任务")
}

// DelSqliteTask 删除sqlite中指定任务
func DelSqliteTask(httpMsg http.ResponseWriter, data *http.Request) {
	tool.Session(httpMsg, "删除sqlite中指定任务")
}
