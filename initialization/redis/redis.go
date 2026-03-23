package redis

import (
	"fmt"
	"planA/initialization/golabl"
	_type "planA/type"
	"time"

	"github.com/go-redis/redis/v8"
)

// Init 初始化Redis连接
// @return error 错误信息
func Init() error {

	// 1. 获取redis配置
	redisConfig := golabl.Config.RedisConfig
	redisClientA, redisErr := NewRedisClient(redisConfig[0])
	if redisErr != nil {
		return fmt.Errorf("初始化 redis %v db%v 失败: %v\n", redisConfig[0].Addr, redisConfig[0].DB, redisErr)
	}
	golabl.RedisDbA = redisClientA

	// Redis B - Redis实例
	redisClientB, redisErr := NewRedisClient(redisConfig[1])
	if redisErr != nil {
		return fmt.Errorf("初始化 redis %v db%v 失败: %v\n", redisConfig[1].Addr, redisConfig[1].DB, redisErr)
	}
	golabl.RedisDbB = redisClientB

	// Redis C - Redis实例
	redisClientC, redisErr := NewRedisClient(redisConfig[2])
	if redisErr != nil {
		return fmt.Errorf("初始化 redis %v db%v 失败: %v\n", redisConfig[2].Addr, redisConfig[2].DB, redisErr)
	}
	golabl.RedisDbC = redisClientC

	// Redis D - Redis实例
	redisClientD, redisErr := NewRedisClient(redisConfig[6])
	if redisErr != nil {
		return fmt.Errorf("初始化 redis %v db%v 失败: %v\n", redisConfig[6].Addr, redisConfig[6].DB, redisErr)
	}
	golabl.RedisDbD = redisClientD

	//设置默认过期时间
	golabl.RedisExp = time.Duration(golabl.Config.Server.RedisExp) * time.Hour
	return nil
}

// NewRedisClient 创建redis 客户端
// @param config redis配置
// @return *redis.Client redis客户端
// @return error 错误信息
func NewRedisClient(config _type.RedisConfig) (*redis.Client, error) {
	ctx := golabl.Ctx
	rdb := redis.NewClient(&redis.Options{
		Addr:               config.Addr,                              // 连接地址
		Password:           config.Password,                          // 密码
		DB:                 config.DB,                                // 数据库
		PoolSize:           config.PoolSize,                          // 连接池大小
		PoolTimeout:        time.Duration(config.PoolTimeout),        // 连接池超时时间
		ReadTimeout:        time.Duration(config.ReadTimeout),        // 读取超时
		WriteTimeout:       time.Duration(config.WriteTimeout),       // 写入超时
		DialTimeout:        time.Duration(config.DialTimeout),        // 连接超时
		IdleTimeout:        time.Duration(config.IdleTimeout),        // 空闲超时
		MinIdleConns:       config.MinIdleConns,                      // 最小空闲连接数
		IdleCheckFrequency: time.Duration(config.IdleCheckFrequency), // 空闲检查频率
		MaxRetries:         config.MaxRetries,                        // 最大重试次数
		MaxRetryBackoff:    time.Duration(config.MaxRetryBackoff),    // 最大重试间隔
		MinRetryBackoff:    time.Duration(config.MinRetryBackoff),    // 最小重试间隔
	})
	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return rdb, err
	}
	return rdb, nil
}
