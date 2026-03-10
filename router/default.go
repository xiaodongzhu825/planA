package router

import (
	"net/http"
	"planA/initialization/golabl"
	"planA/tool"
	_type "planA/type"
)

// DefaultInit 初始化默认路由
func DefaultInit() {
	// 访问根路径时的欢迎页面
	golabl.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tool.JsonResponse(w, http.StatusOK, _type.APIResponse{
			Success: true,
			Message: "任务管理服务已启动",
			Data:    "可用端点: /task/create, /task/goods/*, /task/pause/{id}, /task/resume/{id}, /task/stop/{id}, /task/setTaskBody, /health",
		})
	})
}
