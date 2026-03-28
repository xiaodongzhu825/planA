package golabl

import (
	"context"
	"planA/planB/interfaces"

	planBType "planA/planB/type"
	planAType "planA/type"

	"golang.org/x/time/rate"
)

var (
	Ctx      context.Context        // 全局上下文
	Speed    *rate.Limiter          // 全局令牌桶限速器
	Config   planAType.Config       // 全局配置
	Redis    planBType.Redis        // 全局 Redis
	Task     *planBType.Task        // 全局任务
	Pool     planBType.Pool         // 全局线程池
	Logic    planBType.Logic        // 全局逻辑控制
	Platform interfaces.GoodsTask   // 全局平台对象
	TaskType string                 // 全局任务类型
	MinIo    *planBType.MinIOClient // 全局 MinIO
)

// 任务 body 状态
const (
	BodyStatusSuccess int64 = 1 // 正常
	BodyStatusError   int64 = 2 // 错误
)

// 任务类型
const (
	TaskTypeAddGoodsTask string = "AddGoodsTask" // 添加商品
	TaskTypeGetGoodsTask string = "GetGoodsTask" // 获取商品
	TaskTypeSetGoodsTask string = "SetGoodsTask" // 修改商品
	TaskTypeDelGoodsTask string = "DelGoodsTask" // 删除商品
)

// 错误集
const (
	LastIndexRedisNil            int64 = 10001 // redis 多次读Nil
	LastIndexGoodsMaxRestriction int64 = 11002 // 店铺已达到最大商品限制
	LastIndexFilteWordErr        int64 = 10003 // 过滤关键词异常
)
