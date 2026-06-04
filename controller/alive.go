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
		// v.Times 是原来的计数值
		if v.Times > aliveConfig.Fluent && v.Times < aliveConfig.Slow {
			status = 1
		} else if v.Times >= aliveConfig.Slow {
			status = 2
		}
		ret = append(ret, map[string]interface{}{
			"name":   k,
			"times":  v.Times, // 修改为 v.Times
			"msg":    v.Msg,   // 新增 msg 字段
			"status": status,
		})
	}
	tool.Success(httpMsg, ret)
}
