package _type

import "time"

//配置结构体

// Config 配置结构
type Config struct {
	Server      Server        `json:"server"`
	Alive       Alive         `json:"alive"`
	MysqlConfig MysqlConfig   `json:"mysql_config"`
	PoolConfig  PoolConfig    `json:"pool_config"`
	RedisConfig []RedisConfig `json:"redis_config"`
	PddConfig   PddConfig     `json:"pdd_config"`
	AppBConfig  AppBConfig    `json:"app_b_config"`
	HttpUrl     HttpUrl       `json:"http_url"`
	FileUrl     FileUrl       `json:"file_url"`
}

// Server 服务器配置结构
type Server struct {
	Port         string `json:"port"`
	Filter       int    `json:"filter"`
	ReplaceMark  string `json:"replace_mark"`
	RedisExp     int    `json:"redis_exp"`
	ReadDb       string `json:"read_db"`
	ErrPauseTime int    `json:"err_pause_time"`
}

// Alive 存活状态结构
type Alive struct {
	Fluent int `json:"fluent"`
	Slow   int `json:"slow"`
}

// MysqlConfig Mysql 配置结构
type MysqlConfig struct {
	User              string        `json:"user"`
	Password          string        `json:"password"`
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	DBName            string        `json:"db_name"`
	Loglevel          string        `json:"loglevel"`
	MaxRetryTimes     int           `json:"max_retry_times"`
	BaseRetryInterval time.Duration `json:"base_retry_interval"`
	MaxRetryInterval  time.Duration `json:"max_retry_interval"`
	MaxOpenConns      int           `json:"max_open_conns"`
	MaxIdleConns      int           `json:"max_idle_conns"`
	ConnMaxIdleTime   time.Duration `json:"conn_max_idle_time"`
	ConnMaxLifetime   time.Duration `json:"conn_max_lifetime"`
}

// RedisConfig Redis 配置结构
type RedisConfig struct {
	Addr               string `json:"addr"`
	Password           string `json:"password"`
	DB                 int    `json:"db"`
	PoolSize           int    `json:"pool_size"`
	PoolTimeout        int    `json:"pool_timeout"`
	ReadTimeout        int    `json:"read_timeout"`
	WriteTimeout       int    `json:"write_timeout"`
	DialTimeout        int    `json:"dial_timeout"`
	IdleTimeout        int    `json:"idle_timeout"`
	MinIdleConns       int    `json:"min_idle_conns"`
	IdleCheckFrequency int    `json:"idle_check_frequency"`
	MaxRetries         int    `json:"max_retries"`
	MaxRetryBackoff    int    `json:"max_retry_backoff"`
	MinRetryBackoff    int    `json:"min_retry_backoff"`
}

// PoolConfig 协程池配置结构
type PoolConfig struct {
	Size                 int  `json:"size"`                    // 协程数
	WithExpiryDuration   int  `json:"with_expiry_duration"`    // 过期时间
	WithPreAlloc         bool `json:"with_pre_alloc"`          // 预分配
	WithMaxBlockingTasks int  `json:"with_max_blocking_tasks"` // 最大阻塞任务数
	WithNonblocking      bool `json:"with_nonblocking"`        // 非阻塞
}

// AppBConfig 应用配置结构
type AppBConfig struct {
	AppName string `json:"app_name"` // 应用名称
	AppDir  string `json:"app_dir"`  // 应用目录
}

// PddConfig 拼多多配置
type PddConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// HttpUrl 请求路径
type HttpUrl struct {
	TaskUrl string `json:"task_url"`
}

type FileUrl struct {
	PddDll                    string `json:"pdd_dll"`
	XianYuDll                 string `json:"xian_yu_dll"`
	LogDll                    string `json:"log_dll"`
	ImageDll                  string `json:"image_dll"`
	BFileName                 string `json:"b_file_name"`
	CreateTaskUrl             string `json:"create_task_url"`
	BannedWordSubstitutionUrl string `json:"banned_word_substitution_url"`
	CreateTaskNoticeUrl       string `json:"create_task_notice_url"`
	PddTokenUrl               string `json:"pdd_token_url"`
	DeductionUrl              string `json:"deduction_url"`
}
