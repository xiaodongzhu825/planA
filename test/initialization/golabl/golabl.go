package golabl

import (
	"context"
	TestType "planA/test/type"
	planAType "planA/type"
)

var (
	Ctx      context.Context // 全局上下文
	ApiUrl   string          = "http://127.0.0.1:8080"
	ShopId   string
	ShopType string
	TaskType string
	ImgType  string
	TaskId   string
	Redis    TestType.Redis   // 全局 Redis
	Config   planAType.Config // 全局配置
)
