package golabl

import (
	"context"
	"planA/planE/modules/pdd"
	planAType "planA/type"
)

var (
	Ctx    context.Context  // 全局上下文
	PddDll *pdd.PddDLL      // 全局拼多多 DLL
	Config planAType.Config // 全局配置
)
