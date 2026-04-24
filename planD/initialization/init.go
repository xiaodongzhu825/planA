package initialization

import (
	"context"
	"fmt"
	"planA/planD/initialization/config"
	"planA/planD/initialization/dll"
	"planA/planD/initialization/golabl"
	"planA/planD/initialization/mysql"
)

func Init(taskId string) error {

	//初始化上下文
	if golabl.Ctx == nil {
		golabl.Ctx = context.Background()
	}

	// 初始化配置
	if configErr := config.GetConfigSetToG(); configErr != nil {
		// 配置初始化失败
		return configErr
	}

	// 初始化 任务id
	golabl.TaskId = taskId

	// 初始化 mysql
	if err := mysql.LikeMysqlSetToG(); err != nil {
		// 初始化失败
		return err
	}

	// 初始化 DLL
	if dllErr := dll.GetDllSetToG(); dllErr != nil {
		return fmt.Errorf("初始化DLL失败: %v", dllErr)
	}
	return nil
}
