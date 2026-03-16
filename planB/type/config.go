package _type

import "time"

// Config 配置结构
type Config struct {
	Server      Server
	Alive       Alive
	MysqlConfig MysqlConfig
	PoolConfig  PoolConfig
	RedisConfig []RedisConfig
	PddConfig   PddConfig
	AppBConfig  AppBConfig
	HttpUrl     HttpUrl
	FileUrl     FileUrl
}

// Server 服务器配置结构
type Server struct {
	Port        string
	Filter      int
	ReplaceMark string
}

// Alive 存活状态结构
type Alive struct {
	Fluent int
	Slow   int
}

// MysqlConfig Mysql 配置结构
type MysqlConfig struct {
	User              string
	Password          string
	Host              string
	Port              int
	DBName            string
	Loglevel          string
	MaxRetryTimes     int
	BaseRetryInterval time.Duration
	MaxRetryInterval  time.Duration
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxIdleTime   time.Duration
	ConnMaxLifetime   time.Duration
}

// RedisConfig Redis 配置结构
type RedisConfig struct {
	Addr               string
	Password           string
	DB                 int
	PoolSize           int
	PoolTimeout        int
	ReadTimeout        int
	WriteTimeout       int
	DialTimeout        int
	IdleTimeout        int
	MinIdleConns       int
	IdleCheckFrequency int
	MaxRetries         int
	MaxRetryBackoff    int
	MinRetryBackoff    int
}

// PoolConfig 协程池配置结构
type PoolConfig struct {
	Size                 int  // 协程数
	WithExpiryDuration   int  // 过期时间
	WithPreAlloc         bool // 预分配
	WithMaxBlockingTasks int  // 最大阻塞任务数
	WithNonblocking      bool // 非阻塞
}

// AppBConfig 应用配置结构
type AppBConfig struct {
	AppName string `json:"app_name"` // 应用名称
	AppDir  string `json:"app_dir"`  // 应用目录
}

// PddConfig 拼多多配置
type PddConfig struct {
	ClientId     string
	ClientSecret string
}

// XinayuConfig 闲鱼配置
type XianyuConfig struct {
	ClientId     string
	ClientSecret string
}

// HttpUrl 请求路径
type HttpUrl struct {
	TaskUrl string
}

type FileUrl struct {
	PddDll                    string
	XianYuDll                 string
	LogDll                    string
	ImageDll                  string
	BFileName                 string
	CreateTaskUrl             string
	BannedWordSubstitutionUrl string
	CreateTaskNoticeUrl       string
	PddTokenUrl               string
	DeductionUrl              string
}
