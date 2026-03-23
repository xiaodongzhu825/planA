package initialization

import (
	"context"
	"fmt"
	"planA/planB/initialization/config"
	"planA/planB/initialization/golabl"
	"planA/planB/initialization/platform"
	"planA/planB/initialization/pool"
	"planA/planB/initialization/redis"
	"planA/planB/initialization/speed"
	"planA/planB/initialization/task"
	"planA/planB/initialization/taskType"
	planBType "planA/planB/type"
	planAType "planA/type"
)

// Init 初始化
func Init(taskId string) error {
	//初始化上下文
	if golabl.Ctx == nil {
		golabl.Ctx = context.Background()
	}

	// 初始化限速器
	speed.Init()

	// 初始化配置文件
	if configErr := config.GetConfigSetToG(); configErr != nil {
		return fmt.Errorf("初始化配置文件失败：%v", configErr)
	}

	// 初始化 redis
	if redisErr := redis.LinkRedisSetToG(); redisErr != nil {
		return fmt.Errorf("初始化redis失败: %v", redisErr)
	}

	// 初始化 task
	golabl.Task = &planBType.Task{
		TaskId:     taskId,
		Header:     &planAType.TaskHeader{},
		Footer:     &planAType.TaskFooter{},
		BodyWait:   &planAType.TaskBody{},
		BodyOver:   &planAType.TaskBody{},
		BodyBackup: &planAType.TaskBody{},
	}
	if taskErr := task.GetTaskHeaderAndFooterSetToG(); taskErr != nil {
		return fmt.Errorf("初始化任务失败: %v", taskErr)
	}

	// 初始化 协程池
	if poolErr := pool.CreatePoolToG(); poolErr != nil {
		return fmt.Errorf("初始化协程池失败: %v", poolErr)
	}

	// 初始化平台
	if platformErr := platform.GetPlatformSetToG(); platformErr != nil {
		return fmt.Errorf("初始化平台失败: %v", platformErr)
	}

	// 初始化任务类型
	if taskTypeErr := taskType.GetTaskTypeSetToG(); taskTypeErr != nil {
		return fmt.Errorf("初始化任务类型失败: %v", taskTypeErr)
	}

	return nil
}
