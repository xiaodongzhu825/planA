package interfaces

import (
	planAType "planA/type"
)

// GoodsTask 商品任务接口
type GoodsTask interface {
	// AddGoodsTask 添加商品任务
	AddGoodsTask(bodyWait planAType.TaskBody) (string, error)

	// SetGoodsTask 设置商品任务
	SetGoodsTask() string

	// GetGoodsTask 获取商品任务
	GetGoodsTask() string

	// DelGoodsTask 删除商品任务
	DelGoodsTask() string
}
