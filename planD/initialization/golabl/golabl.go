package golabl

import (
	"context"
	planBType "planA/planB/type"
	"planA/planD/modules/kfz"
	"planA/planD/modules/pdd"

	"gorm.io/gorm"

	planAType "planA/type"
)

var (
	Ctx     context.Context  // 全局上下文
	Config  planAType.Config // 全局配置
	MysqlDb *gorm.DB         // 全局 mysql
	TaskId  string           // 全局任务 ID
	PddDll  *pdd.PddDLL      // 全局拼多多 DLL
	KfzDll  *kfz.KfzDLL      // 全局孔夫子 DLL

	Redis planBType.Redis // 全局 Redis
)
