package serviceAlive

// ServiceStatus 定义服务状态结构体
type ServiceStatus struct {
	Times int    // 次数/状态值
	Msg   string // 消息信息
}

var Service = map[string]ServiceStatus{
	"mysql":          {Times: 0, Msg: ""},
	"redis":          {Times: 0, Msg: ""},
	"sqlite":         {Times: 0, Msg: ""},
	"pdd":            {Times: 0, Msg: ""},
	"通知取出bodyOver接口": {Times: 0, Msg: ""},
	"违禁词替换接口":        {Times: 0, Msg: ""},
	"闲鱼违禁词":          {Times: 0, Msg: ""},
}

// SetServiceAlive 设置服务状态
func SetServiceAlive(key string, times int) {
	if status, ok := Service[key]; ok {
		status.Times = times
		Service[key] = status
	}
}

// SetServiceAliveWithMsg 设置服务状态（带消息）
func SetServiceAliveWithMsg(key string, times int, msg string) {
	Service[key] = ServiceStatus{
		Times: times,
		Msg:   msg,
	}
}

// UpdateServiceTimes 只更新服务次数
func UpdateServiceTimes(key string, times int) {
	if status, ok := Service[key]; ok {
		status.Times = times
		Service[key] = status
	}
}

// UpdateServiceMsg 只更新服务消息
func UpdateServiceMsg(key string, msg string) {
	if status, ok := Service[key]; ok {
		status.Msg = msg
		Service[key] = status
	}
}

// GetServiceStatus 获取服务状态
func GetServiceStatus(key string) (ServiceStatus, bool) {
	status, ok := Service[key]
	return status, ok
}
