package dispatcher

import (
	"fmt"
	"planA/planB/initialization/golabl"
	planAType "planA/type"
)

// Go 调度任务
// @param bodyWait 任务体
// @return string 任务ID
// @return error 错误信息
func Go(bodyWait planAType.TaskBody) (string, error) {
	switch golabl.TaskType {
	case golabl.TaskTypeAddGoodsTask:
		return golabl.Platform.AddGoodsTask(bodyWait) // 添加商品

	//挪到了main方法中执行
	//case "GetGoodsTask":
	//	return golabl.Platform.GetGoodsTask() // 获取商品

	case golabl.TaskTypeSetGoodsTask:

		return golabl.Platform.SetGoodsTask(), nil // 修改商品

	case golabl.TaskTypeOperationGoodsTask:
		return golabl.Platform.OperationGoodsTask(bodyWait) // 操作商品

	case golabl.TaskTypeIncStock:
		return golabl.Platform.IncStockTask(bodyWait) // 增量库存

	default:

		return "", fmt.Errorf("没有此任务类型")
	}
}
