package tool

import "planA/planB/initialization/golabl"

func SteLog(msg string) string {
	platform := GetPlatformName()
	taskTypeName := GetTaskType()
	log := "[任务id：" + golabl.Task.TaskId + "]" + "[店铺名称：" + golabl.Task.Header.ShopName + "]" + "[店铺类型：" + platform + "]" + "[任务类型：" + taskTypeName + "]"
	return log + msg
}
