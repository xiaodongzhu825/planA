package lock

import "sync"

// 用sync.Map替代原生map，天然支持并发安全
var lock sync.Map

// GetLock 获取锁（返回true表示已上锁，false表示未上锁）
func GetLock(key string) bool {
	v, ok := lock.Load(key)
	if !ok {
		return false
	}
	// 断言为bool类型（确保存储的是布尔值）
	locked, ok := v.(bool)
	return ok && locked
}

// SetLock 设置锁（原子操作）
func SetLock(key string) {
	lock.Store(key, true)
}

// DestroyLock 销毁锁（原子操作）
func DestroyLock(key string) {
	lock.Delete(key)
}

// TryLock 尝试加锁（核心：检查+设置原子化）
// 返回true表示加锁成功，false表示已被上锁
func TryLock(key string) bool {
	// LoadOrStore：如果key不存在则存储值，返回false；如果已存在则返回true
	_, loaded := lock.LoadOrStore(key, true)
	// loaded为true表示已上锁，返回false；loaded为false表示加锁成功，返回true
	return !loaded
}
