package taskType

import (
	"errors"
	"fmt"
	"planA/planB/initialization/golabl"
)

// GetTaskTypeSetToG 获取任务类型并保存到全局变量中
// @return error 错误信息
func GetTaskTypeSetToG() error {
	switch golabl.Task.Header.TaskType {
	case 1: //核价发布
		golabl.TaskType = golabl.TaskTypeAddGoodsTask
		return nil
	case 2: //表格发布
		golabl.TaskType = golabl.TaskTypeAddGoodsTask
		return nil
	case 3: //获取商品
		golabl.TaskType = golabl.TaskTypeGetGoodsTask
		return nil
	case 4: //获取拼多多详情商品
		golabl.TaskType = golabl.TaskTypeGetGoodsTask
		return nil
	case 5: //操作商品
		golabl.TaskType = golabl.TaskTypeOperationGoodsTask
		return nil
	case 6: //核价表格发布
		golabl.TaskType = golabl.TaskTypeAddGoodsTask
		return nil
	case 7: //增量库存
		golabl.TaskType = golabl.TaskTypeIncStock
		return nil
	default:
		fmt.Println(golabl.Task.Header.TaskType)
		return errors.New("错误！")
	}
}
