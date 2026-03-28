package _type

import "github.com/go-redis/redis/v8"

// Redis 存储结构
type Redis struct {
	RedisDbA *redis.Client // 任务数据库
	RedisDbB *redis.Client // 出版社数据库
	RedisDbC *redis.Client // 地区数据库
	RedisDbD *redis.Client // 没有书籍的 isbn数据库
}
