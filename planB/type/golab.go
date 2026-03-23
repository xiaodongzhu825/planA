package _type

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/panjf2000/ants/v2"

	planAType "planA/type"
)

// Redis 存储结构
type Redis struct {
	RedisDbA *redis.Client // 任务数据库
	RedisDbB *redis.Client // 出版社数据库
	RedisDbC *redis.Client // 地区数据库
	RedisDbD *redis.Client // 没有书籍的 isbn数据库
}

// Task 任务结构
type Task struct {
	TaskId     string                // 任务ID
	Header     *planAType.TaskHeader // 任务头
	Footer     *planAType.TaskFooter // 任务尾
	BodyWait   *planAType.TaskBody   // 任务等待
	BodyOver   *planAType.TaskBody   // 任务完成
	BodyBackup *planAType.TaskBody   // 任务备份
}

// Pool 线程池结构
type Pool struct {
	Pool *ants.Pool      // 线程池
	Wg   *sync.WaitGroup // 等待组
}

// Logic 逻辑控制结构
type Logic struct {
	TaskIndex      int64 //已读取的 body_wait索引
	RedisNilCon    int64 //连续读出 redisNil 的次数
	ReplaceMarkCon int64 //连续违规词出现的次数
	LastIndex      int64 //记录程序集错误
}
