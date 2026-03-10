package dispatcher

import (
	"fmt"
	"planA/planB/interfaces"
	_type "planA/planB/type"
)

// Go 调度任务
// @param dll 平台对象
// @param types 任务类型
// @param taskHeader 任务头
// @param bodyWait 任务体
// @return string 任务ID
// @return error 错误信息
func Go(dll interfaces.GoodsTask, types string, taskHeader _type.TaskHeader, bodyWait _type.TaskBody) (string, error) {
	//TODO
	switch types {
	case "AddGoodsTask":
		return dll.AddGoodsTask(taskHeader, bodyWait) // 添加商品

	case "GetGoodsTask":

		return dll.GetGoodsTask(), nil // 获取商品

	case "SetGoodsTask":

		return dll.SetGoodsTask(), nil // 修改商品

	case "DelGoodsTask":

		return dll.DelGoodsTask(), nil // 删除商品

	default:

		return "", fmt.Errorf("没有此任务类型")
	}
}
