package pool

import (
	"fmt"
	"planA/planB/initialization/golabl"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

// CreatePoolToG 创建协程池到全局变量中
// @return error 错误信息
func CreatePoolToG() error {
	// 创建协程池
	pool, err := ants.NewPool(
		golabl.Config.PoolConfig.Size,
		ants.WithExpiryDuration(time.Duration(golabl.Config.PoolConfig.WithExpiryDuration)*time.Second), // 过期时间
		ants.WithPreAlloc(golabl.Config.PoolConfig.WithPreAlloc),                                        // 预分配
		ants.WithMaxBlockingTasks(golabl.Config.PoolConfig.WithMaxBlockingTasks),                        // 最大阻塞任务数
		ants.WithNonblocking(golabl.Config.PoolConfig.WithNonblocking),                                  // 非阻塞
		ants.WithPanicHandler(func(p interface{}) { fmt.Printf("panic: %v", p) }),                       // panic 处理
	)
	if err != nil {
		return err
	}
	golabl.Pool.Pool = pool
	golabl.Pool.Wg = &sync.WaitGroup{}
	return nil
}
