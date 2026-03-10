package serviceAlive

var Service = map[string]int{
	"mysql":          0,
	"redis":          0,
	"sqlite":         0,
	"pdd":            0,
	"通知取出bodyOver接口": 0,
	"违禁词替换接口":        0,
}

// SetServiceAlive 设置服务
func SetServiceAlive(key string, times int) {
	Service[key] = times // 现在可以安全写入
}
