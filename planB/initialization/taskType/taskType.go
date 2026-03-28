package taskType

import (
	"errors"
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
	default:
		return errors.New("错误！")
	}
}
