package initialization

import (
	"context"
	"fmt"
	"planA/test/initialization/config"
	"planA/test/initialization/golabl"
	"planA/test/initialization/redis"
)

func Init() error {
	//初始化上下文
	if golabl.Ctx == nil {
		golabl.Ctx = context.Background()
	}
	// 初始化配置文件
	if configErr := config.GetConfigSetToG(); configErr != nil {
		return fmt.Errorf("初始化配置文件失败：%v", configErr)
	}
	// 初始化 redis
	if redisErr := redis.LinkRedisSetToG(); redisErr != nil {
		return fmt.Errorf("初始化redis失败: %v", redisErr)
	}
	return nil
}
