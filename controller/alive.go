package controller

import (
	"net/http"
	"planA/controlState/serviceAlive"
	config2 "planA/initialization/config"
	"planA/tool"
)

// GetServiceAliveList 获取存活状态列表
func GetServiceAliveList(httpMsg http.ResponseWriter, data *http.Request) {

	//获取存活状态列表
	aliveConfig, getAliveConfigErr := config2.GetAliveConfig()
	if getAliveConfigErr != nil {
		tool.Error(httpMsg, getAliveConfigErr.Error(), http.StatusInternalServerError)
		return
	}
	var ret []map[string]interface{}
	alive := serviceAlive.Service
	for k, v := range alive {
		status := 0
		if v > aliveConfig.Fluent && v < aliveConfig.Slow {
			status = 1
		} else if v >= aliveConfig.Slow {
			status = 2
		}
		ret = append(ret, map[string]interface{}{
			"name":   k,
			"times":  v,
			"status": status,
		})
	}
	tool.Session(httpMsg, ret)
}
