package golabl

import (
	_type "planA/planB/type"

	"github.com/go-redis/redis/v8"
)

var (
	RedisClientA *redis.Client
	RedisClientB *redis.Client
	RedisClientC *redis.Client
	MainConfig   *_type.Config
)
