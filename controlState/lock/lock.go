package lock

var lock map[string]string

func init() {
	lock = make(map[string]string) // 初始化
}

// GetLock 获取锁
func GetLock(key string) bool {
	v, ok := lock[key]
	if !ok {
		return false
	}
	return v == "1"
}

// SetLock 设置锁
func SetLock(key string) {
	lock[key] = "1" // 现在可以安全写入
}

// DestroyLock 销毁锁
func DestroyLock(key string) {
	delete(lock, key) // 现在可以安全删除
}
