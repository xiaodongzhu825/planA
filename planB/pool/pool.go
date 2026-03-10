package pool

import (
	"fmt"
	_type "planA/planB/type"
	"time"

	"github.com/panjf2000/ants/v2"
)

// Init 初始化
func Init(poolConfig _type.PoolConfig) (*ants.Pool, error) {
	// 创建协程池
	pool, err := ants.NewPool(
		poolConfig.Size,
		ants.WithExpiryDuration(time.Duration(poolConfig.WithExpiryDuration)*time.Second), // 过期时间
		ants.WithPreAlloc(poolConfig.WithPreAlloc),                                        // 预分配
		ants.WithMaxBlockingTasks(poolConfig.WithMaxBlockingTasks),                        // 最大阻塞任务数
		ants.WithNonblocking(poolConfig.WithNonblocking),                                  // 非阻塞
		ants.WithPanicHandler(func(p interface{}) { fmt.Printf("panic: %v", p) }),         // panic 处理
	)
	// 判断 是否创建成功
	if err != nil {
		return nil, err
	}
	// 返回
	return pool, nil
}
