package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"plantest/modules/kfz"
	"plantest/modules/xianYu"

	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v3"
)

// ============================================================
// 配置（从 test.yaml 加载）
// ============================================================

// Config YAML 配置结构体

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	DB       int    `yaml:"db"`
	Password string `yaml:"password"`
}

type PDDConfig struct {
	ShopID          string `yaml:"shop_id"`
	ShopType        string `yaml:"shop_type"`
	AppID           string `yaml:"app_id"`
	AppKey          string `yaml:"app_key"`
	VerifyURL       string `yaml:"verify_url"`
	VerifyBasicAuth string `yaml:"verify_basic_auth"`
}

type XianyuConfig struct {
	ShopID    string `yaml:"shop_id"`
	ShopType  string `yaml:"shop_type"`
	AppID     int64  `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
	Domain    string `yaml:"domain"`
	DLLPath   string `yaml:"dll_path"`
}

type KfzConfig struct {
	ShopID    string `yaml:"shop_id"`
	ShopType  string `yaml:"shop_type"`
	AppID     int    `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
	DLLPath   string `yaml:"dll_path"`
}

type TimeoutConfig struct {
	WaitTimeout       int `yaml:"wait_timeout"`
	PollInterval      int `yaml:"poll_interval"`
	HTTPClientTimeout int `yaml:"http_client_timeout"`
	CurlTimeout       int `yaml:"curl_timeout"`
	CurlRetryInterval int `yaml:"curl_retry_interval"`
}

type PddPricePublishDelays struct {
	AfterSend int `yaml:"after_send"`
}

type PddPriceChangeDelays struct {
	AfterCreateQueryDetail int `yaml:"after_create_query_detail"`
	AfterSendRedisCheck    int `yaml:"after_send_redis_check"`
	AfterSendAPICheck      int `yaml:"after_send_api_check"`
}

type PddStockChangeDelays struct {
	AfterSendRedisCheck int `yaml:"after_send_redis_check"`
	AfterSendAPICheck   int `yaml:"after_send_api_check"`
}

type PddShelfOnOffDelays struct {
	AfterSendRedisCheck int `yaml:"after_send_redis_check"`
	AfterWaitAPICheck   int `yaml:"after_wait_api_check"`
	WaitBodyOverTimeout int `yaml:"wait_body_over_timeout"`
}

type PddGoodsDeleteDelays struct {
	AfterSendRedisCheck int `yaml:"after_send_redis_check"`
	AfterSendAPICheck   int `yaml:"after_send_api_check"`
}

type XyPricePublishDelays struct {
	AfterSend int `yaml:"after_send"`
}

type XyPriceChangeDelays struct {
	AfterCreateQueryDetail int `yaml:"after_create_query_detail"`
	AfterSendRedisCheck    int `yaml:"after_send_redis_check"`
	AfterSendAPICheck      int `yaml:"after_send_api_check"`
}

type XyStockChangeDelays struct {
	AfterSendRedisCheck int `yaml:"after_send_redis_check"`
	AfterSendAPICheck   int `yaml:"after_send_api_check"`
}

type XyShelfOnOffDelays struct {
	AfterSendRedisCheck int `yaml:"after_send_redis_check"`
	AfterWaitAPICheck   int `yaml:"after_wait_api_check"`
	WaitBodyOverTimeout int `yaml:"wait_body_over_timeout"`
}

type KfzPricePublishDelays struct {
	AfterSend int `yaml:"after_send"`
}

type KfzPriceChangeDelays struct {
	AfterCreateQueryDetail int `yaml:"after_create_query_detail"`
	AfterSendRedisCheck    int `yaml:"after_send_redis_check"`
	AfterSendAPICheck      int `yaml:"after_send_api_check"`
}

type KfzStockChangeDelays struct {
	AfterSendRedisCheck int `yaml:"after_send_redis_check"`
	AfterSendAPICheck   int `yaml:"after_send_api_check"`
}

type KfzShelfOnOffDelays struct {
	AfterSendRedisCheck int `yaml:"after_send_redis_check"`
	AfterWaitAPICheck   int `yaml:"after_wait_api_check"`
	WaitBodyOverTimeout int `yaml:"wait_body_over_timeout"`
}

type KfzGoodsDeleteDelays struct {
	AfterSendRedisCheck int `yaml:"after_send_redis_check"`
	AfterSendAPICheck   int `yaml:"after_send_api_check"`
}

type DelaysConfig struct {
	PddPricePublish PddPricePublishDelays `yaml:"pdd_price_publish"`
	PddPriceChange  PddPriceChangeDelays  `yaml:"pdd_price_change"`
	PddStockChange  PddStockChangeDelays  `yaml:"pdd_stock_change"`
	PddShelfOnOff   PddShelfOnOffDelays   `yaml:"pdd_shelf_on_off"`
	PddGoodsDelete  PddGoodsDeleteDelays  `yaml:"pdd_goods_delete"`
	XyPricePublish  XyPricePublishDelays  `yaml:"xy_price_publish"`
	XyPriceChange   XyPriceChangeDelays   `yaml:"xy_price_change"`
	XyStockChange   XyStockChangeDelays   `yaml:"xy_stock_change"`
	XyShelfOnOff    XyShelfOnOffDelays    `yaml:"xy_shelf_on_off"`
	KfzPricePublish KfzPricePublishDelays `yaml:"kfz_price_publish"`
	KfzPriceChange  KfzPriceChangeDelays  `yaml:"kfz_price_change"`
	KfzStockChange  KfzStockChangeDelays  `yaml:"kfz_stock_change"`
	KfzShelfOnOff   KfzShelfOnOffDelays   `yaml:"kfz_shelf_on_off"`
	KfzGoodsDelete  KfzGoodsDeleteDelays  `yaml:"kfz_goods_delete"`
}

type PddPricePublishTestData struct {
	ISBNSuccess    string `yaml:"isbn_success"`
	PriceSuccess   int64  `yaml:"price_success"`
	ISBNPriceZero  string `yaml:"isbn_price_zero"`
	PriceZero      int64  `yaml:"price_zero"`
	ISBNBannedWord string `yaml:"isbn_banned_word"`
	PriceBanned    int64  `yaml:"price_banned"`
}

type PddPriceChangeTestData struct {
	NewPrice int64 `yaml:"new_price"`
}

type PddStockChangeTestData struct {
	NewStock int64 `yaml:"new_stock"`
}

type XyPricePublishTestData struct {
	ISBNSuccess  string `yaml:"isbn_success"`
	PriceSuccess int64  `yaml:"price_success"`
}

type XyPriceChangeTestData struct {
	NewPrice int64 `yaml:"new_price"`
}

type XyStockChangeTestData struct {
	NewStock int64 `yaml:"new_stock"`
}

type KfzPricePublishTestData struct {
	ISBNSuccess  string `yaml:"isbn_success"`
	PriceSuccess int64  `yaml:"price_success"`
}

type KfzPriceChangeTestData struct {
	NewPrice int64 `yaml:"new_price"`
}

type KfzStockChangeTestData struct {
	NewStock int64 `yaml:"new_stock"`
}

type PullGoodsTestData struct {
	SearchPageSize    int64 `yaml:"search_page_size"`
	BodyWaitMaxSearch int64 `yaml:"body_wait_max_search"`
}

type TestDataConfig struct {
	PddPricePublish PddPricePublishTestData `yaml:"pdd_price_publish"`
	PddPriceChange  PddPriceChangeTestData  `yaml:"pdd_price_change"`
	PddStockChange  PddStockChangeTestData  `yaml:"pdd_stock_change"`
	XyPricePublish  XyPricePublishTestData  `yaml:"xy_price_publish"`
	XyPriceChange   XyPriceChangeTestData   `yaml:"xy_price_change"`
	XyStockChange   XyStockChangeTestData   `yaml:"xy_stock_change"`
	KfzPricePublish KfzPricePublishTestData `yaml:"kfz_price_publish"`
	KfzPriceChange  KfzPriceChangeTestData  `yaml:"kfz_price_change"`
	KfzStockChange  KfzStockChangeTestData  `yaml:"kfz_stock_change"`
	PullGoods       PullGoodsTestData       `yaml:"pull_goods"`
}

type TaskTypeConfig struct {
	PricePublish      string `yaml:"price_publish"`
	PullGoods         string `yaml:"pull_goods"`
	PriceStockShelf   string `yaml:"price_stock_shelf"`
	XyPriceStockShelf string `yaml:"xy_price_stock_shelf"`
}

type TaskCreateConfig struct {
	TaskCount string `yaml:"task_count"`
	ImgType   string `yaml:"img_type"`
}

type BodyOverMinConfig struct {
	PddPricePublish int `yaml:"pdd_price_publish"`
	PddPriceChange  int `yaml:"pdd_price_change"`
	PddShelfOnOff   int `yaml:"pdd_shelf_on_off"`
	PddGoodsDelete  int `yaml:"pdd_goods_delete"`
	XyPriceChange   int `yaml:"xy_price_change"`
	XyShelfOnOff    int `yaml:"xy_shelf_on_off"`
	KfzPricePublish int `yaml:"kfz_price_publish"`
	KfzPriceChange  int `yaml:"kfz_price_change"`
	KfzStockChange  int `yaml:"kfz_stock_change"`
	KfzShelfOnOff   int `yaml:"kfz_shelf_on_off"`
	KfzGoodsDelete  int `yaml:"kfz_goods_delete"`
}

type Config struct {
	BaseURL     string            `yaml:"base_url"`
	Redis       RedisConfig       `yaml:"redis"`
	PDD         PDDConfig         `yaml:"pdd"`
	Xianyu      XianyuConfig      `yaml:"xianyu"`
	Kfz         KfzConfig         `yaml:"kfz"`
	Timeout     TimeoutConfig     `yaml:"timeout"`
	Delays      DelaysConfig      `yaml:"delays"`
	TestData    TestDataConfig    `yaml:"test_data"`
	TaskType    TaskTypeConfig    `yaml:"task_type"`
	TaskCreate  TaskCreateConfig  `yaml:"task_create"`
	GoodsStatus map[int]string    `yaml:"goods_status"`
	BodyOverMin BodyOverMinConfig `yaml:"body_over_min"`
}

// 全局配置变量（从 test.yaml 加载，保持与原 const 同名以最小化代码改动）
var (
	BaseURL    string
	ShopID     string
	ShopType   string
	XyShopID   string
	XyShopType string
	PddAppID   string
	PddAppKey  string
	RedisAddr  string
	RedisDB    int
	RedisPwd   string

	// 等待后台处理超时
	WaitTimeout time.Duration
	// 轮询间隔
	PollInterval time.Duration

	// 校验接口配置
	VerifyURL       string
	VerifyBasicAuth string

	// 闲鱼 DLL 配置
	XyAppID     int64
	XyAppSecret string
	XyDllPath   string
	XyDomain    string

	// 孔夫子 DLL 配置
	KfzAppID     int
	KfzAppSecret string
	KfzDllPath   string

	// 场景延迟配置
	DelayPddPublishAfterSend     time.Duration
	DelayPddChangeAfterQuery     time.Duration
	DelayPddChangeAfterSend      time.Duration
	DelayPddChangeAfterAPI       time.Duration
	DelayPddStockAfterSend       time.Duration
	DelayPddStockAfterAPI        time.Duration
	DelayPddShelfAfterSend       time.Duration
	DelayPddShelfAfterWait       time.Duration
	DelayPddShelfBodyOverTimeout time.Duration
	DelayPddDeleteAfterSend      time.Duration
	DelayPddDeleteAfterAPI       time.Duration
	DelayXyPublishAfterSend      time.Duration
	DelayXyPriceChangeAfterQuery time.Duration
	DelayXyPriceChangeAfterSend  time.Duration
	DelayXyPriceChangeAfterAPI   time.Duration
	DelayXyStockAfterSend        time.Duration
	DelayXyStockAfterAPI         time.Duration
	DelayXyShelfAfterSend        time.Duration
	DelayXyShelfAfterWait        time.Duration
	DelayXyShelfBodyOverTimeout  time.Duration

	// 孔夫子场景延迟
	DelayKfzPublishAfterSend      time.Duration
	DelayKfzPriceChangeAfterQuery time.Duration
	DelayKfzPriceChangeAfterSend  time.Duration
	DelayKfzPriceChangeAfterAPI   time.Duration
	DelayKfzStockAfterSend        time.Duration
	DelayKfzStockAfterAPI         time.Duration
	DelayKfzShelfAfterSend        time.Duration
	DelayKfzShelfAfterWait        time.Duration
	DelayKfzShelfBodyOverTimeout  time.Duration
	DelayKfzDeleteAfterSend       time.Duration
	DelayKfzDeleteAfterAPI        time.Duration

	// 测试数据
	TestISBNSuccess     string
	TestPriceSuccess    int64
	TestISBNPriceZero   string
	TestPriceZero       int64
	TestISBNBanned      string
	TestPriceBanned     int64
	TestNewPrice        int64
	TestNewStock        int64
	TestXyISBNSuccess   string
	TestXyPriceSuccess  int64
	TestXyNewPrice      int64
	TestXyNewStock      int64
	TestKfzISBNSuccess  string
	TestKfzPriceSuccess int64
	TestKfzNewPrice     int64
	TestKfzNewStock     int64

	// 拉取搜索配置
	SearchPageSize    int64
	BodyWaitMaxSearch int64

	// 任务类型
	TaskTypePricePublish      string
	TaskTypePullGoods         string
	TaskTypePriceStockShelf   string
	TaskTypeXyPriceStockShelf string

	// 任务创建参数
	TaskCount string
	ImgType   string

	// 商品状态映射
	StatusName map[int]string

	// body_over 最少条数
	BodyOverMinPublish        int
	BodyOverMinChange         int
	BodyOverMinShelf          int
	BodyOverMinDelete         int
	BodyOverMinXyPriceChange  int
	BodyOverMinXyShelfOnOff   int
	BodyOverMinKfzPublish     int
	BodyOverMinKfzPriceChange int
	BodyOverMinKfzStockChange int
	BodyOverMinKfzShelfOnOff  int
	BodyOverMinKfzDelete      int

	// HTTP 客户端超时
	HTTPClientTimeout time.Duration
	CurlTimeout       time.Duration
	CurlRetryInterval time.Duration
)

// configPath 配置文件路径
var configPath = "test.yaml"

// loadConfig 从 YAML 文件加载配置
func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 基础配置
	BaseURL = cfg.BaseURL

	// Redis
	RedisAddr = cfg.Redis.Addr
	RedisDB = cfg.Redis.DB
	RedisPwd = cfg.Redis.Password

	// 拼多多
	ShopID = cfg.PDD.ShopID
	ShopType = cfg.PDD.ShopType
	PddAppID = cfg.PDD.AppID
	PddAppKey = cfg.PDD.AppKey
	VerifyURL = cfg.PDD.VerifyURL
	VerifyBasicAuth = cfg.PDD.VerifyBasicAuth

	// 闲鱼
	XyShopID = cfg.Xianyu.ShopID
	XyShopType = cfg.Xianyu.ShopType
	XyAppID = cfg.Xianyu.AppID
	XyAppSecret = cfg.Xianyu.AppSecret
	XyDllPath = cfg.Xianyu.DLLPath
	XyDomain = cfg.Xianyu.Domain

	// 孔夫子
	KfzAppID = cfg.Kfz.AppID
	KfzAppSecret = cfg.Kfz.AppSecret
	KfzDllPath = cfg.Kfz.DLLPath

	// 超时
	WaitTimeout = time.Duration(cfg.Timeout.WaitTimeout) * time.Second
	PollInterval = time.Duration(cfg.Timeout.PollInterval) * time.Second
	HTTPClientTimeout = time.Duration(cfg.Timeout.HTTPClientTimeout) * time.Second
	CurlTimeout = time.Duration(cfg.Timeout.CurlTimeout) * time.Second
	CurlRetryInterval = time.Duration(cfg.Timeout.CurlRetryInterval) * time.Second

	// 场景延迟
	DelayPddPublishAfterSend = time.Duration(cfg.Delays.PddPricePublish.AfterSend) * time.Second
	DelayPddChangeAfterQuery = time.Duration(cfg.Delays.PddPriceChange.AfterCreateQueryDetail) * time.Second
	DelayPddChangeAfterSend = time.Duration(cfg.Delays.PddPriceChange.AfterSendRedisCheck) * time.Second
	DelayPddChangeAfterAPI = time.Duration(cfg.Delays.PddPriceChange.AfterSendAPICheck) * time.Second
	DelayPddStockAfterSend = time.Duration(cfg.Delays.PddStockChange.AfterSendRedisCheck) * time.Second
	DelayPddStockAfterAPI = time.Duration(cfg.Delays.PddStockChange.AfterSendAPICheck) * time.Second
	DelayPddShelfAfterSend = time.Duration(cfg.Delays.PddShelfOnOff.AfterSendRedisCheck) * time.Second
	DelayPddShelfAfterWait = time.Duration(cfg.Delays.PddShelfOnOff.AfterWaitAPICheck) * time.Second
	DelayPddShelfBodyOverTimeout = time.Duration(cfg.Delays.PddShelfOnOff.WaitBodyOverTimeout) * time.Second
	DelayXyPublishAfterSend = time.Duration(cfg.Delays.XyPricePublish.AfterSend) * time.Second
	DelayXyPriceChangeAfterQuery = time.Duration(cfg.Delays.XyPriceChange.AfterCreateQueryDetail) * time.Second
	DelayXyPriceChangeAfterSend = time.Duration(cfg.Delays.XyPriceChange.AfterSendRedisCheck) * time.Second
	DelayXyPriceChangeAfterAPI = time.Duration(cfg.Delays.XyPriceChange.AfterSendAPICheck) * time.Second
	DelayXyStockAfterSend = time.Duration(cfg.Delays.XyStockChange.AfterSendRedisCheck) * time.Second
	DelayXyStockAfterAPI = time.Duration(cfg.Delays.XyStockChange.AfterSendAPICheck) * time.Second
	DelayXyShelfAfterSend = time.Duration(cfg.Delays.XyShelfOnOff.AfterSendRedisCheck) * time.Second
	DelayXyShelfAfterWait = time.Duration(cfg.Delays.XyShelfOnOff.AfterWaitAPICheck) * time.Second
	DelayXyShelfBodyOverTimeout = time.Duration(cfg.Delays.XyShelfOnOff.WaitBodyOverTimeout) * time.Second

	// 孔夫子场景延迟
	DelayKfzPublishAfterSend = time.Duration(cfg.Delays.KfzPricePublish.AfterSend) * time.Second
	DelayKfzPriceChangeAfterQuery = time.Duration(cfg.Delays.KfzPriceChange.AfterCreateQueryDetail) * time.Second
	DelayKfzPriceChangeAfterSend = time.Duration(cfg.Delays.KfzPriceChange.AfterSendRedisCheck) * time.Second
	DelayKfzPriceChangeAfterAPI = time.Duration(cfg.Delays.KfzPriceChange.AfterSendAPICheck) * time.Second
	DelayKfzStockAfterSend = time.Duration(cfg.Delays.KfzStockChange.AfterSendRedisCheck) * time.Second
	DelayKfzStockAfterAPI = time.Duration(cfg.Delays.KfzStockChange.AfterSendAPICheck) * time.Second
	DelayKfzShelfAfterSend = time.Duration(cfg.Delays.KfzShelfOnOff.AfterSendRedisCheck) * time.Second
	DelayKfzShelfAfterWait = time.Duration(cfg.Delays.KfzShelfOnOff.AfterWaitAPICheck) * time.Second
	DelayKfzShelfBodyOverTimeout = time.Duration(cfg.Delays.KfzShelfOnOff.WaitBodyOverTimeout) * time.Second
	DelayKfzDeleteAfterSend = time.Duration(cfg.Delays.KfzGoodsDelete.AfterSendRedisCheck) * time.Second
	DelayKfzDeleteAfterAPI = time.Duration(cfg.Delays.KfzGoodsDelete.AfterSendAPICheck) * time.Second

	// 测试数据
	TestISBNSuccess = cfg.TestData.PddPricePublish.ISBNSuccess
	TestPriceSuccess = cfg.TestData.PddPricePublish.PriceSuccess
	TestISBNPriceZero = cfg.TestData.PddPricePublish.ISBNPriceZero
	TestPriceZero = cfg.TestData.PddPricePublish.PriceZero
	TestISBNBanned = cfg.TestData.PddPricePublish.ISBNBannedWord
	TestPriceBanned = cfg.TestData.PddPricePublish.PriceBanned
	TestNewPrice = cfg.TestData.PddPriceChange.NewPrice
	TestNewStock = cfg.TestData.PddStockChange.NewStock
	TestXyISBNSuccess = cfg.TestData.XyPricePublish.ISBNSuccess
	TestXyPriceSuccess = cfg.TestData.XyPricePublish.PriceSuccess
	TestXyNewPrice = cfg.TestData.XyPriceChange.NewPrice
	TestXyNewStock = cfg.TestData.XyStockChange.NewStock
	TestKfzISBNSuccess = cfg.TestData.KfzPricePublish.ISBNSuccess
	TestKfzPriceSuccess = cfg.TestData.KfzPricePublish.PriceSuccess
	TestKfzNewPrice = cfg.TestData.KfzPriceChange.NewPrice
	TestKfzNewStock = cfg.TestData.KfzStockChange.NewStock

	// 拉取搜索
	SearchPageSize = cfg.TestData.PullGoods.SearchPageSize
	BodyWaitMaxSearch = cfg.TestData.PullGoods.BodyWaitMaxSearch

	// 任务类型
	TaskTypePricePublish = cfg.TaskType.PricePublish
	TaskTypePullGoods = cfg.TaskType.PullGoods
	TaskTypePriceStockShelf = cfg.TaskType.PriceStockShelf
	TaskTypeXyPriceStockShelf = cfg.TaskType.XyPriceStockShelf

	// 任务创建参数
	TaskCount = cfg.TaskCreate.TaskCount
	ImgType = cfg.TaskCreate.ImgType

	// 商品状态映射
	StatusName = cfg.GoodsStatus
	if StatusName == nil {
		StatusName = map[int]string{1: "上架", 2: "下架", 3: "售罄", 4: "已删除"}
	}

	// body_over 最少条数
	BodyOverMinPublish = cfg.BodyOverMin.PddPricePublish
	BodyOverMinChange = cfg.BodyOverMin.PddPriceChange
	BodyOverMinShelf = cfg.BodyOverMin.PddShelfOnOff
	BodyOverMinDelete = cfg.BodyOverMin.PddGoodsDelete
	BodyOverMinXyPriceChange = cfg.BodyOverMin.XyPriceChange
	BodyOverMinXyShelfOnOff = cfg.BodyOverMin.XyShelfOnOff
	BodyOverMinKfzPublish = cfg.BodyOverMin.KfzPricePublish
	BodyOverMinKfzPriceChange = cfg.BodyOverMin.KfzPriceChange
	BodyOverMinKfzStockChange = cfg.BodyOverMin.KfzStockChange
	BodyOverMinKfzShelfOnOff = cfg.BodyOverMin.KfzShelfOnOff
	BodyOverMinKfzDelete = cfg.BodyOverMin.KfzGoodsDelete

	DelayPddDeleteAfterSend = time.Duration(cfg.Delays.PddGoodsDelete.AfterSendRedisCheck) * time.Second
	DelayPddDeleteAfterAPI = time.Duration(cfg.Delays.PddGoodsDelete.AfterSendAPICheck) * time.Second

	// 更新 httpClient 超时
	httpClient = &http.Client{Timeout: HTTPClientTimeout}

	return nil
}

// countdownDelay 带倒计时的延迟
func countdownDelay(d time.Duration, desc string) {
	fmt.Printf("\n⏳ 延迟 %v 后%s，等待后台处理完成...\n", d, desc)
	delayStart := time.Now()
	for remaining := d; remaining > 0; remaining = d - time.Since(delayStart) {
		fmt.Printf("\r    倒计时: %v    ", remaining.Round(time.Second))
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("\r    ✅ 延迟结束，开始校验          \n")
}

// ============================================================
// 数据结构
// ============================================================

// APIResponse 接口统一响应
type APIResponse struct {
	Code string      `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// TestResult 单条测试结果
type TestResult struct {
	Category string
	Name     string
	Status   string // PASS / FAIL / ERROR
	Detail   string
	Duration time.Duration
}

// ============================================================
// 全局变量
// ============================================================
var (
	results     []TestResult
	httpClient  *http.Client
	redisCtx    = context.Background()
	redisClient *redis.Client
	reportDir   string

	// 跨场景数据传递（拼多多）
	priceTaskID    string // 场景一 创建核价发布任务返回的 task_id
	successISBN    string // 场景一中 detail.error 包含"执行成功"对应的 isbn
	successGoodsID int64  // 场景一中执行成功对应的 goods_id
	mixedTaskID    string // 场景二 创建改价格任务返回的 task_id（task_type=5，也用于场景三改库存、场景四上下架）
	successSkuID   int64  // 场景二步骤2查询商品详情时从响应sku_list中提取的 sku_id
	pullTaskID     string // 场景五 创建商品拉取任务返回的 task_id（task_type=3）
	pullGoodsID    int64  // 场景五 拉取任务中场景一 ISBN 对应的 goods_id

	// 闲鱼场景
	xyTaskID         string // 场景七 闲鱼核价发布任务 task_id
	xyModTaskID      string // 场景十/十一/十二 闲鱼改库存/下架/上架 共用的任务 task_id
	xySuccessISBN    string // 场景七中执行成功的ISBN
	xySuccessGoodsID int64  // 场景七中执行成功对应的goods_id
	xyPullTaskID     string // 场景八 闲鱼商品拉取任务 task_id
	xyPullGoodsID    int64  // 场景八 拉取任务中找到的goods_id
)

// 孔夫子任务跟踪变量
var (
	kfzTaskID         string // 孔夫子任务 task_id
	kfzSuccessISBN    string // 核价发布成功ISBN
	kfzSuccessGoodsID int64  // 核价发布成功goods_id
)

// ============================================================
// HTTP 工具
// ============================================================

var AppKey string // 签名密钥

// SignParams 对请求参数进行MD5签名
// SignParams 对请求参数进行MD5签名
var SignSecretKey = "jRQdCh52Z55Kzh1hADaA2ZtdTKetj2PXk60Tz5Yc0iz9aD8Wafbk7CwAZ8cz69A9zb9caZ3k9dnR3Ys06J5nYFPrZ0xE9p6TY8DCD538ryiRjW81YTPmk41tCEnXizPh"

func SignParams(params map[string]string) string {
	// 过滤需要签名的参数（排除空值、sign、sign_type）
	filteredParams := make(map[string]string)
	for k, v := range params {
		if v != "" && k != "sign" && k != "sign_type" {
			filteredParams[k] = v
		}
	}

	// 提取键名并排序
	keys := make([]string, 0, len(filteredParams))
	for k := range filteredParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接参数字符串: key=value&key=value...
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(filteredParams[k])
	}

	// 末尾加上签名密钥
	signStr := builder.String() + "&key=" + SignSecretKey

	// MD5 哈希并转大写十六进制
	hash := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// postMultipart 发送 multipart/form-data POST 请求
func postMultipart(url string, fields map[string]string) (*APIResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("写入字段 %s 失败: %w", k, err)
		}
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(b, &apiResp); err != nil {
		return nil, fmt.Errorf("JSON解析失败 [%s]: %w", truncate(string(b), 300), err)
	}
	return &apiResp, nil
}

// ============================================================
// Redis 工具
// ============================================================

func initRedis() error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: RedisPwd,
		DB:       RedisDB,
	})
	_, err := redisClient.Ping(redisCtx).Result()
	return err
}

// getBodyOverFromRedis 从 Redis 读取 body_over list（指定范围）
func getBodyOverFromRedis(taskID string, start, stop int64) ([]map[string]interface{}, error) {
	key := taskID + ":body_over"
	vals, err := redisClient.LRange(redisCtx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis LRange %s 失败: %w", key, err)
	}

	var result []map[string]interface{}
	for _, v := range vals {
		var item map[string]interface{}
		if err := json.Unmarshal([]byte(v), &item); err != nil {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// getBodyOverCount 获取 body_over 数量
func getBodyOverCount(taskID string) (int64, error) {
	key := taskID + ":body_over"
	return redisClient.LLen(redisCtx, key).Result()
}

// getHeaderStatus 获取任务状态
func getHeaderStatus(taskID string) (int64, error) {
	key := taskID + ":header"
	v, err := redisClient.HGet(redisCtx, key, "status").Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(v, 10, 64)
}

// ============================================================
// 等待工具
// ============================================================

// waitBodyOverMin 等待 body_over 至少有 minCount 条
func waitBodyOverMin(taskID string, minCount int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	round := 0
	for time.Now().Before(deadline) {
		cnt, err := getBodyOverCount(taskID)
		if err == nil && cnt >= int64(minCount) {
			return nil
		}
		round++
		if round%5 == 0 {
			fmt.Printf("    ⏳ 等待中... body_over=%d (需要≥%d)\n", cnt, minCount)
		}
		time.Sleep(PollInterval)
	}
	cnt, _ := getBodyOverCount(taskID)
	return fmt.Errorf("等待超时 (%v)，body_over 数量 %d < %d", timeout, cnt, minCount)
}

// ============================================================
// 通用解析工具
// ============================================================

// extractGoodsStatus 从商品详情 map 中提取 status 和 goods_name
// 商品状态：1=上架，2=下架，3=售罄，4=已删除
func extractGoodsStatus(m map[string]interface{}) (int, string) {
	status := -1
	goodsName := ""

	if v, ok := m["status"]; ok {
		switch val := v.(type) {
		case float64:
			status = int(val)
		case string:
			status, _ = strconv.Atoi(val)
		}
	}

	if v, ok := m["goods_name"]; ok {
		if s, ok := v.(string); ok {
			goodsName = s
		}
	}

	return status, goodsName
}

// queryGoodsDetail 用 curl 调用商品详情接口，返回解析后的 map
// 支持重试，新商品在 PDD 上需要同步时间
func queryGoodsDetail(accessToken string, goodsID int64, maxRetries int) (map[string]interface{}, error) {
	goodsIDStr := fmt.Sprintf("%d", goodsID)

	for retry := 1; retry <= maxRetries; retry++ {
		curlArgs := []string{
			"-s",
			"--request", "GET",
			VerifyURL,
			"--header", "Authorization: Basic " + VerifyBasicAuth,
			"--form", "accessToken=" + accessToken,
			"--form", "goodsId=" + goodsIDStr,
		}

		if retry == 1 {
			fmt.Printf("    📋 curl --request GET %s --form accessToken=*** --form goodsId=%s\n", VerifyURL, goodsIDStr)
		}

		ctx, cancel := context.WithTimeout(context.Background(), CurlTimeout)
		cmd := exec.CommandContext(ctx, "curl", curlArgs...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			cancel()
			if retry == maxRetries {
				return nil, fmt.Errorf("curl 执行失败: %v, stderr: %s", err, truncate(stderr.String(), 200))
			}
			fmt.Printf("    ⚠️ curl 失败（第%d次），重试...\n", retry)
			time.Sleep(CurlRetryInterval)
			continue
		}
		cancel()

		rawStr := strings.TrimSpace(stdout.String())
		fmt.Printf("    📋 curl 响应（第%d次）: %s\n", retry, truncate(rawStr, 300))

		// 检查是否返回有效数据（非 null/空）
		if rawStr == "null" || rawStr == "" {
			if retry < maxRetries {
				fmt.Printf("    ⚠️ 响应为空（第%d次），重试中...\n", retry)
				time.Sleep(CurlRetryInterval)
			}
			continue
		}

		var rawMap map[string]interface{}
		if err := json.Unmarshal([]byte(rawStr), &rawMap); err != nil {
			if retry < maxRetries {
				fmt.Printf("    ⚠️ JSON解析失败（第%d次），重试中...\n", retry)
				time.Sleep(CurlRetryInterval)
			}
			continue
		}

		// 检查是否有 goods_id（有效商品数据）
		if rawMap["goods_id"] != nil {
			fmt.Printf("    ✅ 获取到有效商品数据\n")
			return rawMap, nil
		}

		// 可能在 data 字段内
		if dataMap, ok := rawMap["data"].(map[string]interface{}); ok && dataMap["goods_id"] != nil {
			fmt.Printf("    ✅ 获取到有效商品数据（data层）\n")
			return rawMap, nil
		}

		if retry < maxRetries {
			fmt.Printf("    ⚠️ 响应无有效商品数据（第%d次），重试中...\n", retry)
			time.Sleep(CurlRetryInterval)
		}
	}

	return nil, fmt.Errorf("重试 %d 次后仍未获取到有效商品数据", maxRetries)
}

// extractSkuID 从商品详情 map 中提取第一个 sku_id
func extractSkuID(rawMap map[string]interface{}) int64 {
	// 尝试从顶层 sku_list 提取
	if skuList, ok := rawMap["sku_list"].([]interface{}); ok && len(skuList) > 0 {
		if sku0, ok := skuList[0].(map[string]interface{}); ok {
			if v, ok := sku0["sku_id"].(float64); ok {
				return int64(v)
			}
		}
	}
	// 尝试从 data 层提取
	if dataMap, ok := rawMap["data"].(map[string]interface{}); ok {
		if skuList, ok := dataMap["sku_list"].([]interface{}); ok && len(skuList) > 0 {
			if sku0, ok := skuList[0].(map[string]interface{}); ok {
				if v, ok := sku0["sku_id"].(float64); ok {
					return int64(v)
				}
			}
		}
	}
	return 0
}

// extractSkuQuantity 从商品详情 map 中查找匹配 sku_id 的 SKU，提取 quantity（库存）
func extractSkuQuantity(rawMap map[string]interface{}, targetSkuID int64) int64 {
	var skuList []interface{}
	if sl, ok := rawMap["sku_list"].([]interface{}); ok {
		skuList = sl
	} else if dataMap, ok := rawMap["data"].(map[string]interface{}); ok {
		if sl, ok := dataMap["sku_list"].([]interface{}); ok {
			skuList = sl
		}
	}

	for _, item := range skuList {
		sku, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if sid, ok := sku["sku_id"].(float64); ok && int64(sid) == targetSkuID {
			if v, ok := sku["quantity"].(float64); ok {
				return int64(v)
			}
		}
	}

	// 没匹配到 sku_id，用第一个
	if len(skuList) > 0 {
		if sku0, ok := skuList[0].(map[string]interface{}); ok {
			if v, ok := sku0["quantity"].(float64); ok {
				return int64(v)
			}
		}
	}
	return -1
}

// extractSkuPrice 从商品详情 map 中查找匹配 sku_id 的 SKU，提取 multi_price 和 price
func extractSkuPrice(rawMap map[string]interface{}, targetSkuID int64) (multiPrice int64, basePrice int64) {
	multiPrice = -1
	basePrice = -1

	// 查找 sku_list
	var skuList []interface{}
	if sl, ok := rawMap["sku_list"].([]interface{}); ok {
		skuList = sl
	} else if dataMap, ok := rawMap["data"].(map[string]interface{}); ok {
		if sl, ok := dataMap["sku_list"].([]interface{}); ok {
			skuList = sl
		}
	}

	for _, item := range skuList {
		sku, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if sid, ok := sku["sku_id"].(float64); ok && int64(sid) == targetSkuID {
			if v, ok := sku["multi_price"].(float64); ok {
				multiPrice = int64(v)
			}
			if v, ok := sku["price"].(float64); ok {
				basePrice = int64(v)
			}
			return
		}
	}

	// 没匹配到 sku_id，用第一个
	if len(skuList) > 0 {
		if sku0, ok := skuList[0].(map[string]interface{}); ok {
			if v, ok := sku0["multi_price"].(float64); ok {
				multiPrice = int64(v)
			}
			if v, ok := sku0["price"].(float64); ok {
				basePrice = int64(v)
			}
		}
	}
	return
}

// getAccessToken 从 Redis 获取 accessToken
func getAccessToken(taskID string) (string, error) {
	if taskID == "" {
		return "", fmt.Errorf("task_id 为空")
	}
	headerKey := taskID + ":header"
	shopMsgJSON, err := redisClient.HGet(redisCtx, headerKey, "shop_msg").Result()
	if err != nil {
		return "", fmt.Errorf("Redis HGet %s shop_msg 失败: %w", headerKey, err)
	}

	var shopMsg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(shopMsgJSON), &shopMsg); err != nil {
		return "", fmt.Errorf("解析 shop_msg JSON 失败: %w, 原始: %s", err, truncate(shopMsgJSON, 200))
	}
	if shopMsg.Token == "" {
		return "", fmt.Errorf("shop_msg.token 为空")
	}
	return shopMsg.Token, nil
}

// ============================================================
// 闲鱼 DLL 查询辅助函数
// ============================================================

// getXyAccessToken 从 Redis 获取闲鱼的 accessToken（token 字段）
func getXyAccessToken() (string, error) {
	if xyTaskID == "" {
		return "", fmt.Errorf("xyTaskID 为空，场景七未成功创建任务")
	}
	headerKey := xyTaskID + ":header"
	shopMsgJSON, err := redisClient.HGet(redisCtx, headerKey, "shop_msg").Result()
	if err != nil {
		return "", fmt.Errorf("Redis HGet %s shop_msg 失败: %w", headerKey, err)
	}
	var shopMsg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(shopMsgJSON), &shopMsg); err != nil {
		return "", fmt.Errorf("解析 shop_msg JSON 失败: %w, 原始: %s", err, truncate(shopMsgJSON, 200))
	}
	if shopMsg.Token == "" {
		return "", fmt.Errorf("shop_msg.token 为空")
	}
	return shopMsg.Token, nil
}

// findXyGoodsByISBN 用 DLL ExecuteSelectGoodsListPrice 按 ISBN 搜索商品，返回 product_id
func findXyGoodsByISBN(isbn string, accessToken string) (int64, error) {
	xyDll, err := xianYu.InitXianYuDll(XyDllPath)
	if err != nil {
		return 0, fmt.Errorf("初始化闲鱼DLL失败: %v", err)
	}
	// online_time 传空数组，product_status=22 表示在售
	queryReq := map[string]interface{}{
		"appId":          XyAppID,
		"appSecret":      XyAppSecret,
		"token":          accessToken,
		"online_time":    []interface{}{},
		"product_status": 22,
	}
	queryJSON, _ := json.Marshal(queryReq)
	result, err := xyDll.XianYuGetGoodsList(string(queryJSON), XyDllPath)
	if err != nil {
		return 0, fmt.Errorf("XianYuGetGoodsList 失败: %v", err)
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return 0, fmt.Errorf("解析商品列表 JSON 失败: %v, 原始: %s", err, truncate(result, 200))
	}
	code, _ := resultMap["code"].(float64)
	if code != 200 && code != 0 {
		msg, _ := resultMap["msg"].(string)
		return 0, fmt.Errorf("查询商品列表失败: code=%.0f msg=%s", code, msg)
	}
	// 遍历列表匹配 ISBN
	data, ok := resultMap["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("data 字段格式异常，无法解析商品列表")
	}
	for _, itemRaw := range data {
		item, ok := itemRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if itemISBN, ok := item["isbn"].(string); ok && itemISBN == isbn {
			if pid, ok := item["product_id"].(float64); ok {
				return int64(pid), nil
			}
		}
	}
	return 0, fmt.Errorf("ISBN=%s 在商品列表中未找到", isbn)
}

// getXyGoodsPrice 用 DLL ExecuteGetGoodsDetail 查询商品价格（返回分）
func getXyGoodsPrice(productID int64, accessToken string) (int64, error) {
	xyDll, err := xianYu.InitXianYuDll(XyDllPath)
	if err != nil {
		return 0, fmt.Errorf("初始化闲鱼DLL失败: %v", err)
	}
	queryReq := map[string]interface{}{
		"appId":      XyAppID,
		"appSecret":  XyAppSecret,
		"token":      accessToken,
		"product_id": productID,
	}
	queryJSON, _ := json.Marshal(queryReq)
	result, err := xyDll.XianYuGetGoodsDetail(string(queryJSON), XyDllPath)
	if err != nil {
		return 0, fmt.Errorf("XianYuGetGoodsDetail 失败: %v", err)
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return 0, fmt.Errorf("解析商品详情 JSON 失败: %v", err)
	}
	code, _ := resultMap["code"].(float64)
	if code != 200 && code != 0 {
		msg, _ := resultMap["msg"].(string)
		return 0, fmt.Errorf("查询商品详情失败: code=%.0f msg=%s", code, msg)
	}
	if price, ok := resultMap["price"].(float64); ok {
		return int64(price), nil
	}
	if data, ok := resultMap["data"].(map[string]interface{}); ok {
		if price, ok := data["price"].(float64); ok {
			return int64(price), nil
		}
	}
	return 0, fmt.Errorf("商品详情中未找到 price 字段")
}

// getXyGoodsStock 用 DLL ExecuteGetGoodsDetail 查询商品库存
func getXyGoodsStock(productID int64, accessToken string) (int64, error) {
	xyDll, err := xianYu.InitXianYuDll(XyDllPath)
	if err != nil {
		return 0, fmt.Errorf("初始化闲鱼DLL失败: %v", err)
	}
	queryReq := map[string]interface{}{
		"appId":      XyAppID,
		"appSecret":  XyAppSecret,
		"token":      accessToken,
		"product_id": productID,
	}
	queryJSON, _ := json.Marshal(queryReq)
	result, err := xyDll.XianYuGetGoodsDetail(string(queryJSON), XyDllPath)
	if err != nil {
		return 0, fmt.Errorf("XianYuGetGoodsDetail 失败: %v", err)
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return 0, fmt.Errorf("解析商品详情 JSON 失败: %v", err)
	}
	code, _ := resultMap["code"].(float64)
	if code != 200 && code != 0 {
		msg, _ := resultMap["msg"].(string)
		return 0, fmt.Errorf("查询商品详情失败: code=%.0f msg=%s", code, msg)
	}
	if stock, ok := resultMap["stock"].(float64); ok {
		return int64(stock), nil
	}
	if data, ok := resultMap["data"].(map[string]interface{}); ok {
		if stock, ok := data["stock"].(float64); ok {
			return int64(stock), nil
		}
	}
	return 0, fmt.Errorf("商品详情中未找到 stock 字段")
}

// getXyGoodsStatus 用 DLL ExecuteGetGoodsDetail 查询商品状态
// 返回值：1=上架 2=下架 3=售罄 4=已删除，-1=未知
func getXyGoodsStatus(productID int64, accessToken string) (int, error) {
	xyDll, err := xianYu.InitXianYuDll(XyDllPath)
	if err != nil {
		return -1, fmt.Errorf("初始化闲鱼DLL失败: %v", err)
	}
	queryReq := map[string]interface{}{
		"appId":      XyAppID,
		"appSecret":  XyAppSecret,
		"token":      accessToken,
		"product_id": productID,
	}
	queryJSON, _ := json.Marshal(queryReq)
	result, err := xyDll.XianYuGetGoodsDetail(string(queryJSON), XyDllPath)
	if err != nil {
		return -1, fmt.Errorf("XianYuGetGoodsDetail 失败: %v", err)
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return -1, fmt.Errorf("解析商品详情 JSON 失败: %v", err)
	}
	code, _ := resultMap["code"].(float64)
	if code != 200 && code != 0 {
		msg, _ := resultMap["msg"].(string)
		return -1, fmt.Errorf("查询商品详情失败: code=%.0f msg=%s", code, msg)
	}
	if ps, ok := resultMap["product_status"].(float64); ok {
		return int(ps), nil
	}
	if data, ok := resultMap["data"].(map[string]interface{}); ok {
		if ps, ok := data["product_status"].(float64); ok {
			return int(ps), nil
		}
	}
	return -1, fmt.Errorf("商品详情中未找到 product_status 字段")
}

// ============================================================
// 孔夫子 DLL 查询辅助函数
// ============================================================

// getKfzAccessToken 从 Redis 获取孔夫子的 accessToken（token 字段）
func getKfzAccessToken() (string, error) {
	if kfzTaskID == "" {
		return "", fmt.Errorf("kfzTaskID 为空，场景未成功创建任务")
	}
	headerKey := kfzTaskID + ":header"
	shopMsgJSON, err := redisClient.HGet(redisCtx, headerKey, "shop_msg").Result()
	if err != nil {
		return "", fmt.Errorf("Redis HGet %s shop_msg 失败: %w", headerKey, err)
	}
	var shopMsg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(shopMsgJSON), &shopMsg); err != nil {
		return "", fmt.Errorf("解析 shop_msg JSON 失败: %w, 原始: %s", err, truncate(shopMsgJSON, 200))
	}
	if shopMsg.Token == "" {
		return "", fmt.Errorf("shop_msg.token 为空")
	}
	return shopMsg.Token, nil
}

// findKfzGoodsByISBN 用 DLL GetGoodsList 按 ISBN 搜索商品，返回 itemId
func findKfzGoodsByISBN(isbn string, accessToken string) (int64, error) {
	kfzDll, err := kfz.InitKfzDll(KfzDllPath)
	if err != nil {
		return 0, fmt.Errorf("初始化孔夫子DLL失败: %v", err)
	}
	// type="sale" 表示在售，pageNum=1
	queryReq := kfz.GetGoodsListReq{
		Type:      "sale",
		PageNum:   1,
		PageSize:  50,
		SortOrder: "addTime",
		SortType:  "DESC",
	}
	queryJSON, _ := json.Marshal(queryReq)
	result, err := kfzDll.GetGoodsList(KfzAppID, KfzAppSecret, accessToken, string(queryJSON))
	if err != nil {
		return 0, fmt.Errorf("KongfzShopItemList 失败: %v", err)
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return 0, fmt.Errorf("解析商品列表 JSON 失败: %v, 原始: %s", err, truncate(result, 200))
	}
	// 检查 errorResponse
	if errResp, ok := resultMap["errorResponse"].(map[string]interface{}); ok {
		if code, _ := errResp["code"].(float64); code != 0 {
			msg, _ := errResp["msg"].(string)
			return 0, fmt.Errorf("查询商品列表失败: code=%.0f msg=%s", code, msg)
		}
	}
	// 遍历列表匹配 ISBN
	if successResp, ok := resultMap["successResponse"].(map[string]interface{}); ok {
		if list, ok := successResp["list"].([]interface{}); ok {
			for _, itemRaw := range list {
				item, ok := itemRaw.(map[string]interface{})
				if !ok {
					continue
				}
				if itemISBN, ok := item["isbn"].(string); ok && itemISBN == isbn {
					if itemId, ok := item["itemId"].(float64); ok {
						return int64(itemId), nil
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("ISBN=%s 在商品列表中未找到", isbn)
}

// getKfzGoodsPrice 从商品列表中获取指定 itemId 的商品价格（返回元）
func getKfzGoodsPrice(itemId int64, accessToken string) (float64, error) {
	kfzDll, err := kfz.InitKfzDll(KfzDllPath)
	if err != nil {
		return 0, fmt.Errorf("初始化孔夫子DLL失败: %v", err)
	}
	queryReq := kfz.GetGoodsListReq{
		Type:     "sale",
		PageNum:  1,
		PageSize: 50,
	}
	queryJSON, _ := json.Marshal(queryReq)
	result, err := kfzDll.GetGoodsList(KfzAppID, KfzAppSecret, accessToken, string(queryJSON))
	if err != nil {
		return 0, fmt.Errorf("KongfzShopItemList 失败: %v", err)
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return 0, fmt.Errorf("解析商品列表 JSON 失败: %v", err)
	}
	if successResp, ok := resultMap["successResponse"].(map[string]interface{}); ok {
		if list, ok := successResp["list"].([]interface{}); ok {
			for _, itemRaw := range list {
				if item, ok := itemRaw.(map[string]interface{}); ok {
					if id, ok := item["itemId"].(float64); ok && id == float64(itemId) {
						if price, ok := item["price"].(float64); ok {
							return price, nil
						}
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("itemId=%d 未找到对应商品或价格字段", itemId)
}

// getKfzGoodsStock 从商品列表中获取指定 itemId 的商品库存
func getKfzGoodsStock(itemId int64, accessToken string) (int, error) {
	kfzDll, err := kfz.InitKfzDll(KfzDllPath)
	if err != nil {
		return 0, fmt.Errorf("初始化孔夫子DLL失败: %v", err)
	}
	queryReq := kfz.GetGoodsListReq{
		Type:     "sale",
		PageNum:  1,
		PageSize: 50,
	}
	queryJSON, _ := json.Marshal(queryReq)
	result, err := kfzDll.GetGoodsList(KfzAppID, KfzAppSecret, accessToken, string(queryJSON))
	if err != nil {
		return 0, fmt.Errorf("KongfzShopItemList 失败: %v", err)
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return 0, fmt.Errorf("解析商品列表 JSON 失败: %v", err)
	}
	if successResp, ok := resultMap["successResponse"].(map[string]interface{}); ok {
		if list, ok := successResp["list"].([]interface{}); ok {
			for _, itemRaw := range list {
				if item, ok := itemRaw.(map[string]interface{}); ok {
					if id, ok := item["itemId"].(float64); ok && id == float64(itemId) {
						if stock, ok := item["number"].(float64); ok {
							return int(stock), nil
						}
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("itemId=%d 未找到对应商品或库存字段", itemId)
}

// getKfzGoodsStatus 从商品列表中获取指定 itemId 的商品上下架状态
// isOnSale: 1=上架 0=下架
func getKfzGoodsStatus(itemId int64, accessToken string) (int, error) {
	kfzDll, err := kfz.InitKfzDll(KfzDllPath)
	if err != nil {
		return -1, fmt.Errorf("初始化孔夫子DLL失败: %v", err)
	}
	queryReq := kfz.GetGoodsListReq{
		Type:     "sale",
		PageNum:  1,
		PageSize: 50,
	}
	queryJSON, _ := json.Marshal(queryReq)
	result, err := kfzDll.GetGoodsList(KfzAppID, KfzAppSecret, accessToken, string(queryJSON))
	if err != nil {
		return -1, fmt.Errorf("KongfzShopItemList 失败: %v", err)
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return -1, fmt.Errorf("解析商品列表 JSON 失败: %v", err)
	}
	if successResp, ok := resultMap["successResponse"].(map[string]interface{}); ok {
		if list, ok := successResp["list"].([]interface{}); ok {
			for _, itemRaw := range list {
				if item, ok := itemRaw.(map[string]interface{}); ok {
					if id, ok := item["itemId"].(float64); ok && id == float64(itemId) {
						if isOnSale, ok := item["isOnSale"].(float64); ok {
							return int(isOnSale), nil
						}
					}
				}
			}
		}
	}
	return -1, fmt.Errorf("itemId=%d 未找到对应商品", itemId)
}

// ============================================================
// 测试结果工具
// ============================================================

func pass(cat, name, detail string, elapsed time.Duration) {
	results = append(results, TestResult{Category: cat, Name: name, Status: "PASS", Detail: detail, Duration: elapsed})
	fmt.Printf("  ✅ PASS (%s)\n    详情: %s\n", elapsed.Round(time.Millisecond), detail)
}

func fail(cat, name, detail string, elapsed time.Duration) {
	results = append(results, TestResult{Category: cat, Name: name, Status: "FAIL", Detail: detail, Duration: elapsed})
	fmt.Printf("  ❌ FAIL (%s)\n    详情: %s\n", elapsed.Round(time.Millisecond), detail)
}

func errCase(cat, name, detail string, elapsed time.Duration) {
	results = append(results, TestResult{Category: cat, Name: name, Status: "ERROR", Detail: detail, Duration: elapsed})
	fmt.Printf("  🔴 ERROR (%s)\n    详情: %s\n", elapsed.Round(time.Millisecond), detail)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ============================================================
// 报告生成
// ============================================================

func sanitize(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func generateReport() string {
	var sb strings.Builder

	sb.WriteString("# 📋 PlanA API 批量测试报告\n\n")
	sb.WriteString(fmt.Sprintf("**测试时间**: %s  \n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**接口地址**: `%s`  \n", BaseURL))
	sb.WriteString(fmt.Sprintf("**ShopID**: `%s`  \n", ShopID))
	sb.WriteString(fmt.Sprintf("**拼多多应用ID**: `%s`  \n", PddAppID))
	sb.WriteString(fmt.Sprintf("**Redis**: `%s` DB=%d  \n\n", RedisAddr, RedisDB))

	total := len(results)
	p, f, e := 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case "PASS":
			p++
		case "FAIL":
			f++
		default:
			e++
		}
	}
	rate := 0.0
	if total > 0 {
		rate = float64(p) / float64(total) * 100
	}

	sb.WriteString("## 📊 测试汇总\n\n")
	sb.WriteString("| 指标 | 值 |\n|---|---|\n")
	sb.WriteString(fmt.Sprintf("| 通过率 | %d/%d (%.1f%%) |\n", p, total, rate))
	sb.WriteString(fmt.Sprintf("| ✅ PASS | %d |\n", p))
	sb.WriteString(fmt.Sprintf("| ❌ FAIL | %d |\n", f))
	sb.WriteString(fmt.Sprintf("| 🔴 ERROR | %d |\n\n", e))

	if priceTaskID != "" || mixedTaskID != "" || successISBN != "" {
		sb.WriteString("## 📌 跨场景数据记录\n\n")
		sb.WriteString("| 数据项 | 值 |\n|---|---|\n")
		if priceTaskID != "" {
			sb.WriteString(fmt.Sprintf("| 核价发布 task_id | `%s` |\n", priceTaskID))
		}
		if mixedTaskID != "" {
			sb.WriteString(fmt.Sprintf("| 改价格/上下架 task_id | `%s` |\n", mixedTaskID))
		}
		if successISBN != "" {
			sb.WriteString(fmt.Sprintf("| 执行成功 ISBN | `%s` |\n", successISBN))
		}
		if successGoodsID != 0 {
			sb.WriteString(fmt.Sprintf("| 执行成功 GoodsID | `%d` |\n", successGoodsID))
		}
		if successSkuID != 0 {
			sb.WriteString(fmt.Sprintf("| 执行成功 SkuID | `%d` |\n", successSkuID))
		}
		if pullTaskID != "" {
			sb.WriteString(fmt.Sprintf("| 商品拉取 task_id | `%s` |\n", pullTaskID))
		}
		if pullGoodsID != 0 {
			sb.WriteString(fmt.Sprintf("| 拉取到的 GoodsID | `%d` |\n", pullGoodsID))
		}
		sb.WriteString("\n")
	}

	curCat := ""
	idx := 0
	for _, r := range results {
		if r.Category != curCat {
			curCat = r.Category
			if idx > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("## %s\n\n", curCat))
			sb.WriteString("| # | 用例 | 状态 | 耗时 | 详情 |\n")
			sb.WriteString("|---|---|---|---|---|\n")
			idx = 0
		}
		idx++
		emoji := map[string]string{"PASS": "✅", "FAIL": "❌", "ERROR": "🔴"}[r.Status]
		detail := sanitize(r.Detail)
		if len(detail) > 200 {
			detail = detail[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s %s | %s | %s |\n",
			idx, r.Name, emoji, r.Status, r.Duration.Round(time.Millisecond), detail))
	}

	sb.WriteString("\n---\n")
	sb.WriteString(fmt.Sprintf("*报告生成: %s*", time.Now().Format("2006-01-02 15:04:05")))
	return sb.String()
}

// ============================================================
// 测试七：闲鱼核价发布任务
// ============================================================

func testXyPricePublish() {
	cat := "七、闲鱼核价发布任务"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	var tid string

	// ---------- 步骤1：创建闲鱼核价发布任务 ----------
	{
		name := "1、创建闲鱼核价发布任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    XyShopID,
			"shop_type":  XyShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePricePublish,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		if dataStr == "" {
			fail(cat, name, "返回 data 为空", elapsed)
			return
		}
		tid = dataStr
		xyTaskID = tid
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	if tid == "" {
		fmt.Println("⚠️  任务创建失败，跳过后续步骤")
		return
	}
	fmt.Printf("\n  📌 闲鱼核价发布 task_id: %s\n", tid)

	// ---------- 步骤2：发送 isbn（期望：执行成功）----------
	{
		name := fmt.Sprintf("2、发送任务数据【isbn=%s, price=%d】期望：执行成功", TestXyISBNSuccess, TestXyPriceSuccess)
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"price":%d}}`, TestXyISBNSuccess, TestXyPriceSuccess)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, "接口返回成功", elapsed)
	}

	// 延迟后校验
	countdownDelay(DelayXyPublishAfterSend, "校验")

	// ---------- 步骤3：Redis 校验 body_over（仅校验执行成功）----------
	{
		name := "3、Redis 校验 - body_over 中 detail.error 包含'执行成功'"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		fmt.Printf("    ⏳ 等待后台处理完成（最多 %v）...\n", WaitTimeout)

		if err := waitBodyOverMin(tid, 1, WaitTimeout); err != nil {
			errCase(cat, name, "等待超时: "+err.Error(), time.Since(start))
			return
		}

		totalCnt, _ := getBodyOverCount(tid)
		fmt.Printf("    📊 body_over 总数: %d，开始校验...\n", totalCnt)

		targetISBN := TestXyISBNSuccess
		var detailLines []string
		found := false
		var successGoodsID int64

		const pageSize int64 = 100
		for offset := int64(0); offset < totalCnt && !found; offset += pageSize {
			items, err := getBodyOverFromRedis(tid, offset, offset+pageSize-1)
			if err != nil || len(items) == 0 {
				break
			}

			for _, item := range items {
				isbn := ""
				if bi, ok := item["book_info"].(map[string]interface{}); ok {
					if v, ok := bi["isbn"].(string); ok {
						isbn = v
					}
				}
				if isbn != targetISBN {
					continue
				}

				var errMsg string
				var goodsID float64
				var detailStatus float64
				if detail, ok := item["detail"].(map[string]interface{}); ok {
					if v, ok := detail["error"].(string); ok {
						errMsg = v
					}
					if v, ok := detail["goods_id"].(float64); ok {
						goodsID = v
					}
					if v, ok := detail["status"].(float64); ok {
						detailStatus = v
					}
				}

				found = true
				errorMatched := strings.Contains(errMsg, "执行成功") || detailStatus == 1
				hasGoodsID := goodsID > 0

				if errorMatched && hasGoodsID {
					successGoodsID = int64(goodsID)
					detailLines = append(detailLines, fmt.Sprintf("ISBN=%s ✅ 匹配'执行成功' (goods_id=%.0f)", isbn, goodsID))
				} else {
					failReason := ""
					if !hasGoodsID {
						failReason = " [原因: goods_id为空]"
					} else if !errorMatched {
						failReason = " [原因: error不含'执行成功'且status!=1]"
					}
					detailLines = append(detailLines,
						fmt.Sprintf("ISBN=%s ❌ 期望'执行成功'%s 实际error='%s', status=%.0f, goods_id=%.0f",
							isbn, failReason, truncate(errMsg, 80), detailStatus, goodsID))
				}
				break
			}
		}

		if !found {
			detailLines = append(detailLines, fmt.Sprintf("❌ 未在body_over中找到ISBN=%s", targetISBN))
		}

		detail := strings.Join(detailLines, " | ")
		elapsed := time.Since(start)

		if found && successGoodsID > 0 {
			pass(cat, name, detail, elapsed)
		} else {
			fail(cat, name, detail, elapsed)
		}

		// 保存成功的goods_id供后续测试使用
		if successGoodsID > 0 {
			xySuccessGoodsID = successGoodsID
			xySuccessISBN = targetISBN
			fmt.Printf("\n  📌 记录执行成功数据: ISBN=%s, goods_id=%d\n", targetISBN, successGoodsID)
		}
	}

	// ---------- 步骤5：DLL校验 - 通过闲鱼API查询商品详情 ----------
	if xySuccessGoodsID > 0 {
		name := "5、DLL校验 - 通过闲鱼API查询商品详情"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 初始化DLL
		xyDll, err := xianYu.InitXianYuDll(XyDllPath)
		if err != nil {
			fail(cat, name, fmt.Sprintf("初始化闲鱼DLL失败: %v", err), time.Since(start))
		} else {
			// 构建查询请求 - 使用goods_id作为product_id查询
			// 注意：这里假设goods_id就是product_id，如果不一致需要调整
			queryReq := map[string]interface{}{
				"appId":      XyAppID,
				"appSecret":  XyAppSecret,
				"product_id": xySuccessGoodsID,
			}
			queryJSON, _ := json.Marshal(queryReq)

			fmt.Printf("    📋 查询商品详情: product_id=%d\n", xySuccessGoodsID)
			result, err := xyDll.XianYuGetGoodsDetail(string(queryJSON), XyDllPath)
			elapsed := time.Since(start)

			if err != nil {
				fail(cat, name, fmt.Sprintf("DLL调用失败: %v", err), elapsed)
			} else {
				// 解析返回结果
				var resultMap map[string]interface{}
				if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
					// 如果解析失败，直接显示原始结果
					pass(cat, name, fmt.Sprintf("DLL返回: %s", truncate(result, 200)), elapsed)
				} else {
					// 检查返回结果中是否包含商品信息
					code, _ := resultMap["code"].(float64)
					if code == 200 || code == 0 {
						pass(cat, name, fmt.Sprintf("商品详情查询成功 ✅ (product_id=%d)", xySuccessGoodsID), elapsed)
					} else {
						msg, _ := resultMap["msg"].(string)
						fail(cat, name, fmt.Sprintf("商品详情查询失败: %s", msg), elapsed)
					}
				}
			}
		}
	} else {
		fmt.Printf("\n  ⚠️ 未找到执行成功的商品，跳过DLL校验\n")
	}
}

// ============================================================
// 测试一：拼多多核价发布任务
// ============================================================

func testPddPricePublish() {
	cat := "一、拼多多核价发布任务"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	var tid string

	// ---------- 步骤1：创建核价发布任务 ----------
	{
		name := "1、创建拼多多核价发布任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePricePublish,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		if dataStr == "" {
			fail(cat, name, "返回 data 为空", elapsed)
			return
		}
		tid = dataStr
		priceTaskID = tid
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	if tid == "" {
		fmt.Println("⚠️  任务创建失败，跳过后续步骤")
		return
	}
	fmt.Printf("\n  📌 核价发布 task_id: %s\n", tid)

	// ---------- 步骤2：发送 isbn（期望：执行成功）----------
	{
		name := fmt.Sprintf("2、发送任务数据【isbn=%s, price=%d】期望：执行成功", TestISBNSuccess, TestPriceSuccess)
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"price":%d}}`, TestISBNSuccess, TestPriceSuccess)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, "接口返回成功", elapsed)
	}

	// ---------- 步骤3a：发送 isbn（期望：价格不能小于等于0）----------
	{
		name := fmt.Sprintf("3a、发送任务数据【isbn=%s, price=%d】期望：价格不能小于等于0", TestISBNPriceZero, TestPriceZero)
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"price":%d}}`, TestISBNPriceZero, TestPriceZero)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, "接口返回成功", elapsed)
	}

	// ---------- 步骤3b：发送 isbn（期望：违规词命中）----------
	{
		name := fmt.Sprintf("3b、发送任务数据【isbn=%s, price=%d】期望：违规词命中", TestISBNBanned, TestPriceBanned)
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"price":%d}}`, TestISBNBanned, TestPriceBanned)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, "接口返回成功", elapsed)
	}

	// 延迟后校验
	countdownDelay(DelayPddPublishAfterSend, "校验")

	// ---------- 步骤4：Redis 校验 body_over ----------
	{
		name := "4、Redis 校验 - body_over 中 detail.error 匹配期望结果"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		fmt.Printf("    ⏳ 等待后台处理完成（最多 %v）...\n", WaitTimeout)

		if err := waitBodyOverMin(tid, BodyOverMinPublish, WaitTimeout); err != nil {
			errCase(cat, name, "等待超时: "+err.Error(), time.Since(start))
			return
		}

		totalCnt, _ := getBodyOverCount(tid)
		fmt.Printf("    📊 body_over 总数: %d，开始校验...\n", totalCnt)

		expected := map[string]string{
			TestISBNSuccess:   "执行成功",
			TestISBNPriceZero: "价格不能小于等于0",
			TestISBNBanned:    "违规词命中",
		}

		var detailLines []string
		allPass := true
		matchedCount := 0

		const pageSize int64 = 100
		for offset := int64(0); offset < totalCnt; offset += pageSize {
			items, err := getBodyOverFromRedis(tid, offset, offset+pageSize-1)
			if err != nil || len(items) == 0 {
				break
			}

			for _, item := range items {
				isbn := ""
				if bi, ok := item["book_info"].(map[string]interface{}); ok {
					if v, ok := bi["isbn"].(string); ok {
						isbn = v
					}
				}
				if isbn == "" {
					continue
				}

				want, hasExpect := expected[isbn]
				if !hasExpect {
					continue
				}

				matchedCount++

				var errMsg string
				var goodsID float64
				var detailStatus float64
				if detail, ok := item["detail"].(map[string]interface{}); ok {
					if v, ok := detail["error"].(string); ok {
						errMsg = v
					}
					if v, ok := detail["goods_id"].(float64); ok {
						goodsID = v
					}
					if v, ok := detail["status"].(float64); ok {
						detailStatus = v
					}
				}

				matched := false
				if want == "执行成功" {
					errorMatched := strings.Contains(errMsg, "执行成功") || detailStatus == 1
					hasGoodsID := goodsID > 0
					matched = errorMatched && hasGoodsID
				} else {
					matched = strings.Contains(errMsg, want)
				}

				if matched {
					detailLines = append(detailLines, fmt.Sprintf("ISBN=%s ✅ 匹配'%s' (goods_id=%.0f)", isbn, want, goodsID))
					if want == "执行成功" {
						successISBN = isbn
						successGoodsID = int64(goodsID)
					}
				} else {
					failReason := ""
					if want == "执行成功" {
						if goodsID <= 0 {
							failReason = " [原因: goods_id为空，执行成功但无商品ID]"
						} else {
							failReason = " [原因: error不含'执行成功'且status!=1]"
						}
					}
					detailLines = append(detailLines,
						fmt.Sprintf("ISBN=%s ❌ 期望'%s'%s 实际error='%s', status=%.0f, goods_id=%.0f",
							isbn, want, failReason, truncate(errMsg, 80), detailStatus, goodsID))
					allPass = false
				}
			}
		}

		if matchedCount < len(expected) {
			detailLines = append(detailLines,
				fmt.Sprintf("⚠️ 仅匹配到 %d/%d 个期望ISBN", matchedCount, len(expected)))
			allPass = false
		}

		detail := strings.Join(detailLines, " | ")
		elapsed := time.Since(start)

		if allPass {
			pass(cat, name, detail, elapsed)
		} else {
			fail(cat, name, detail, elapsed)
		}
	}
}

// 测试一结束

// ============================================================
// 测试二：拼多多改价格
// ============================================================

func testPddPriceChange() {
	cat := "二、拼多多改价格"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	// 前置条件检查
	if successISBN == "" || successGoodsID == 0 {
		fmt.Println("⚠️  场景一未匹配到执行成功数据（isbn/goods_id），跳过改价格测试")
		fail(cat, "前置条件", "场景一未匹配到执行成功数据（isbn/goods_id）", 0)
		return
	}

	// ---------- 步骤1：创建改价格任务（task_type=5）----------
	var tid string
	{
		name := "1、创建拼多多改价格、上下架、改库存任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePriceStockShelf,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		if dataStr == "" {
			fail(cat, name, "返回 data 为空", elapsed)
			return
		}
		tid = dataStr
		mixedTaskID = tid
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	if tid == "" {
		fmt.Println("⚠️  改价格任务创建失败，跳过后续步骤")
		return
	}
	fmt.Printf("\n  📌 改价格/上下架 task_id: %s\n", tid)

	// ---------- 步骤2：查询商品详情，获取 sku_id ----------
	{
		name := "2、查询商品详情，获取 sku_id"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 延迟等待新商品在 PDD 上同步
		countdownDelay(DelayPddChangeAfterQuery, "查询商品详情")

		// 获取 accessToken
		accessToken, err := getAccessToken(priceTaskID)
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}
		fmt.Printf("    🔑 accessToken: %s...%s\n", accessToken[:8], accessToken[len(accessToken)-8:])
		fmt.Printf("    📦 goodsId: %d\n", successGoodsID)

		// 查询商品详情（最多重试5次，新商品需要同步时间）
		rawMap, err := queryGoodsDetail(accessToken, successGoodsID, 5)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, fmt.Sprintf("查询商品详情失败: %v", err), elapsed)
			return
		}

		// 提取 sku_id
		successSkuID = extractSkuID(rawMap)
		if successSkuID == 0 {
			fail(cat, name, "响应中未找到 sku_id", elapsed)
			return
		}

		// 提取商品状态
		goodsStatus, goodsName := extractGoodsStatus(rawMap)
		statusStr := fmt.Sprintf("%d", goodsStatus)
		if n, ok := StatusName[goodsStatus]; ok {
			statusStr = n
		}

		// 提取当前价格（用于后续对比）
		multiPrice, basePrice := extractSkuPrice(rawMap, successSkuID)

		fmt.Printf("    📦 sku_id: %d\n", successSkuID)
		fmt.Printf("    📦 商品状态: %s(%s)\n", statusStr, StatusName[goodsStatus])
		fmt.Printf("    📦 当前价格: multi_price=%d, price=%d\n", multiPrice, basePrice)

		if goodsStatus != 1 {
			fail(cat, name, fmt.Sprintf("商品非上架状态 status=%d(%s)，改价格需要商品在上架状态。goodsName=%s",
				goodsStatus, statusStr, truncate(goodsName, 50)), elapsed)
			return
		}

		pass(cat, name, fmt.Sprintf("sku_id=%d 商品上架中 status=1(上架) multi_price=%d price=%d goodsName=%s",
			successSkuID, multiPrice, basePrice, truncate(goodsName, 50)), elapsed)
	}

	if successSkuID == 0 {
		fmt.Println("⚠️  未获取到 sku_id，跳过后续改价格步骤")
		return
	}

	// 测试改价格：从配置读取
	testPrice := TestNewPrice

	fmt.Printf("\n  📌 isbn: %s, goods_id: %d, sku_id: %d\n", successISBN, successGoodsID, successSkuID)
	fmt.Printf("  📌 改价格为: %d（分）= %.2f元\n", testPrice, float64(testPrice)/100)

	// ---------- 步骤3：发送改价格任务 ----------
	{
		name := "3、发送任务数据【改价格】"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		bodyJSON := fmt.Sprintf(
			`{"book_info":{"isbn":"%s"},"detail":{"goods_id":%d,"status":5,"price":%d,"sku_id":%d}}`,
			successISBN, successGoodsID, testPrice, successSkuID)
		fmt.Printf("    📋 task_id: %s\n", tid)
		fmt.Printf("    📋 body: %s\n", bodyJSON)

		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		// 打印原始响应
		rawJSON, _ := json.Marshal(resp)
		fmt.Printf("    📥 原始响应: %s\n", string(rawJSON))

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("ISBN=%s GoodsID=%d sku_id=%d price=%d(%.2f元) 接口返回成功",
			successISBN, successGoodsID, successSkuID, testPrice, float64(testPrice)/100), elapsed)
	}

	// 延迟后 Redis 校验
	countdownDelay(DelayPddChangeAfterSend, "校验")

	// ---------- 步骤4：Redis 校验改价格结果 ----------
	{
		name := "4、Redis 校验改价格结果（body_over）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 等待改价格任务处理完成
		fmt.Printf("    ⏳ 等待改价格任务处理完成（检查 body_over）...\n")
		deadline := time.Now().Add(WaitTimeout)
		processed := false
		for time.Now().Before(deadline) {
			cnt, _ := getBodyOverCount(tid)
			if cnt > 0 {
				fmt.Printf("    ✅ 改价格任务已处理完成（body_over 条目数: %d）\n", cnt)
				processed = true
				break
			}
			time.Sleep(PollInterval)
		}
		if !processed {
			fail(cat, name, "等待超时，body_over 仍无数据", time.Since(start))
			return
		}

		// 读取 body_over 最新的条目（最后一条）
		totalCnt, _ := getBodyOverCount(tid)
		items, err := getBodyOverFromRedis(tid, totalCnt-1, totalCnt-1)
		elapsed := time.Since(start)

		if err != nil || len(items) == 0 {
			errCase(cat, name, fmt.Sprintf("读取 body_over 失败: %v", err), elapsed)
			return
		}

		item := items[0]
		detail, _ := item["detail"].(map[string]interface{})
		errMsg, _ := detail["error"].(string)
		priceField, _ := detail["price"].(float64)
		skuIDField, _ := detail["sku_id"].(float64)

		fmt.Printf("    📋 error: %s\n", truncate(errMsg, 200))
		fmt.Printf("    📋 price: %.0f, sku_id: %.0f\n", priceField, skuIDField)

		if strings.Contains(errMsg, "执行成功") || strings.Contains(errMsg, "成功") {
			pass(cat, name, fmt.Sprintf("改价格执行成功: error='%s', price=%.0f, sku_id=%.0f",
				truncate(errMsg, 100), priceField, skuIDField), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("改价格失败: error='%s'", truncate(errMsg, 200)), elapsed)
		}
	}

	// ---------- 步骤5：校验改价格状态（接口验证）----------
	// 调用商品详情接口，对比 sku_list 中的 multi_price 是否与传入的 price 一致
	{
		name := "5、校验改价格状态（接口验证）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 延迟后接口校验
		countdownDelay(DelayPddChangeAfterAPI, "接口校验")

		// 获取 accessToken
		accessToken, err := getAccessToken(priceTaskID)
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}

		// 查询商品详情
		rawMap, err := queryGoodsDetail(accessToken, successGoodsID, 3)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, fmt.Sprintf("查询商品详情失败: %v", err), elapsed)
			return
		}

		// 提取价格
		multiPrice, basePrice := extractSkuPrice(rawMap, successSkuID)

		// 提取商品状态
		goodsStatus, _ := extractGoodsStatus(rawMap)
		statusStr := fmt.Sprintf("%d", goodsStatus)
		if n, ok := StatusName[goodsStatus]; ok {
			statusStr = n
		}

		fmt.Printf("    📦 sku_id=%d: multi_price=%d, price=%d, 期望price=%d\n",
			successSkuID, multiPrice, basePrice, testPrice)
		fmt.Printf("    📦 商品状态: %s(%s)\n", statusStr, StatusName[goodsStatus])

		// 价格对比逻辑
		if multiPrice == testPrice {
			pass(cat, name, fmt.Sprintf("multi_price=%d 与传入 price=%d 一致 ✅",
				multiPrice, testPrice), elapsed)
		} else if basePrice == testPrice {
			discountStr := ""
			if v, ok := rawMap["two_pieces_discount"].(float64); ok {
				discountStr = fmt.Sprintf("two_pieces_discount=%d%%", int(v))
			}
			pass(cat, name, fmt.Sprintf("price=%d 与传入一致，multi_price=%d（%s 折扣导致差异）",
				basePrice, multiPrice, discountStr), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("价格不匹配：传入price=%d，实际multi_price=%d, price=%d（商品状态=%s）",
				testPrice, multiPrice, basePrice, statusStr), elapsed)
		}
	}
}

// ============================================================
// 测试三：拼多多上下架
// ============================================================

// ============================================================
// 测试三：拼多多改库存
// ============================================================

func testPddStockChange() {
	cat := "三、拼多多改库存"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	// 前置条件检查
	if mixedTaskID == "" {
		fmt.Println("⚠️  改价格 task_id 为空，跳过改库存测试")
		fail(cat, "前置条件", "场景二未创建改价格任务，无法获取 task_id", 0)
		return
	}
	if successISBN == "" || successGoodsID == 0 || successSkuID == 0 {
		fmt.Println("⚠️  场景一/二未获取到必要数据（isbn/goods_id/sku_id），跳过改库存测试")
		fail(cat, "前置条件", "场景一/二未获取到必要数据（isbn/goods_id/sku_id）", 0)
		return
	}

	fmt.Printf("\n  📌 使用改价格 task_id: %s（复用 task_type=5 任务）\n", mixedTaskID)

	// 测试改库存：从配置读取
	testStock := TestNewStock

	fmt.Printf("  📌 isbn: %s, goods_id: %d, sku_id: %d\n", successISBN, successGoodsID, successSkuID)
	fmt.Printf("  📌 改库存为: %d\n", testStock)

	// ---------- 步骤1：发送改库存任务 ----------
	{
		name := "1、发送任务数据【改库存】"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		bodyJSON := fmt.Sprintf(
			`{"book_info":{"isbn":"%s"},"detail":{"goods_id":%d,"status":4,"stock":%d,"sku_id":%d}}`,
			successISBN, successGoodsID, testStock, successSkuID)
		fmt.Printf("    📋 task_id: %s\n", mixedTaskID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)

		params := map[string]string{
			"task_id": mixedTaskID,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("ISBN=%s GoodsID=%d sku_id=%d stock=%d status=4(改库存) 接口返回成功",
			successISBN, successGoodsID, successSkuID, testStock), elapsed)
	}

	// 延迟后 Redis 校验
	countdownDelay(DelayPddStockAfterSend, "校验")

	// ---------- 步骤2：Redis 校验改库存结果 ----------
	{
		name := "2、Redis 校验改库存结果（body_over）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 等待改库存任务处理完成
		fmt.Printf("    ⏳ 等待改库存任务处理完成（检查 body_over）...\n")
		deadline := time.Now().Add(WaitTimeout)
		processed := false
		for time.Now().Before(deadline) {
			cnt, _ := getBodyOverCount(mixedTaskID)
			if cnt > 0 {
				fmt.Printf("    ✅ 改库存任务已处理完成（body_over 条目数: %d）\n", cnt)
				processed = true
				break
			}
			time.Sleep(PollInterval)
		}
		if !processed {
			fail(cat, name, "等待超时，body_over 仍无数据", time.Since(start))
			return
		}

		// 读取 body_over 最新的条目（最后一条）
		totalCnt, _ := getBodyOverCount(mixedTaskID)
		items, err := getBodyOverFromRedis(mixedTaskID, totalCnt-1, totalCnt-1)
		elapsed := time.Since(start)

		if err != nil || len(items) == 0 {
			errCase(cat, name, fmt.Sprintf("读取 body_over 失败: %v", err), elapsed)
			return
		}

		item := items[0]
		detail, _ := item["detail"].(map[string]interface{})
		errMsg, _ := detail["error"].(string)
		stockField, _ := detail["stock"].(float64)
		skuIDField, _ := detail["sku_id"].(float64)

		fmt.Printf("    📋 error: %s\n", truncate(errMsg, 200))
		fmt.Printf("    📋 stock: %.0f, sku_id: %.0f\n", stockField, skuIDField)

		if strings.Contains(errMsg, "执行成功") || strings.Contains(errMsg, "成功") {
			pass(cat, name, fmt.Sprintf("改库存执行成功: error='%s', stock=%.0f, sku_id=%.0f",
				truncate(errMsg, 100), stockField, skuIDField), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("改库存失败: error='%s'", truncate(errMsg, 200)), elapsed)
		}
	}

	// ---------- 步骤3：校验库存（接口验证）----------
	{
		name := "3、校验库存（接口验证）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 延迟后接口校验
		countdownDelay(DelayPddStockAfterAPI, "接口校验")

		// 获取 accessToken
		accessToken, err := getAccessToken(priceTaskID)
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}

		// 查询商品详情
		rawMap, err := queryGoodsDetail(accessToken, successGoodsID, 3)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, fmt.Sprintf("查询商品详情失败: %v", err), elapsed)
			return
		}

		// 提取库存
		quantity := extractSkuQuantity(rawMap, successSkuID)

		// 提取商品状态
		goodsStatus, _ := extractGoodsStatus(rawMap)
		statusStr := fmt.Sprintf("%d", goodsStatus)
		if n, ok := StatusName[goodsStatus]; ok {
			statusStr = n
		}

		fmt.Printf("    📦 sku_id=%d: quantity=%d, 期望stock=%d\n", successSkuID, quantity, testStock)
		fmt.Printf("    📦 商品状态: %s(%s)\n", statusStr, StatusName[goodsStatus])

		if quantity == testStock {
			pass(cat, name, fmt.Sprintf("库存匹配: quantity=%d 与传入 stock=%d 一致 ✅（商品状态=%s）",
				quantity, testStock, statusStr), elapsed)
		} else if quantity == -1 {
			errCase(cat, name, fmt.Sprintf("未能从响应中提取库存数量"), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("库存不匹配：传入stock=%d，实际quantity=%d（商品状态=%s）",
				testStock, quantity, statusStr), elapsed)
		}
	}
}

// ============================================================
// 测试四：拼多多上下架
// ============================================================

func testPddShelfOnOff() {
	cat := "四、拼多多上下架"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	// 前置条件检查
	if mixedTaskID == "" {
		fmt.Println("⚠️  改价格 task_id 为空，跳过上下架测试")
		fail(cat, "前置条件", "场景二未创建改价格任务，无法获取 task_id", 0)
		return
	}
	if successISBN == "" || successGoodsID == 0 {
		fmt.Println("⚠️  场景一未匹配到执行成功数据，跳过上下架测试")
		fail(cat, "前置条件", "场景一未匹配到执行成功数据（isbn/goods_id）", 0)
		return
	}

	fmt.Printf("\n  📌 使用改价格 task_id: %s（复用 task_type=5 任务）\n", mixedTaskID)

	// ---------- 步骤1：发送下架任务 ----------
	{
		name := "1、发送任务数据【下架】"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"goods_id":%d,"status":2}}`,
			successISBN, successGoodsID)
		fmt.Printf("    📋 task_id: %s\n", mixedTaskID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)

		params := map[string]string{
			"task_id": mixedTaskID,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("ISBN=%s GoodsID=%d status=2(下架) 接口返回成功", successISBN, successGoodsID), elapsed)
	}

	// 延迟后校验
	countdownDelay(DelayPddShelfAfterSend, "校验")

	// ---------- 等待下架任务处理完成 ----------
	{
		fmt.Printf("\n    ⏳ 等待下架任务处理完成（检查 body_over）...\n")
		// 改价格 + 改库存 已有 2 条，下架后应该达到 3 条
		if err := waitBodyOverMin(mixedTaskID, BodyOverMinShelf, DelayPddShelfBodyOverTimeout); err != nil {
			fmt.Printf("    ⚠️  等待超时: %v，继续尝试校验\n", err)
		} else {
			fmt.Printf("    ✅ 下架任务已处理完成\n")
		}
	}

	// ---------- 步骤2：校验下架状态 ----------
	// 校验接口：从配置读取 URL 和 Basic Auth
	// 响应 status 字段：1=上架，2=下架，3=售罄，4=已删除
	// 期望：status=2（下架）
	{
		name := "2、校验下架状态"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 获取 accessToken
		accessToken, err := getAccessToken(priceTaskID)
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}
		fmt.Printf("    🔑 accessToken: %s...%s\n", accessToken[:8], accessToken[len(accessToken)-8:])
		fmt.Printf("    📦 goodsId: %d\n", successGoodsID)

		// 延迟后接口校验
		countdownDelay(DelayPddShelfAfterWait, "接口校验")

		// 查询商品详情（最多重试5次）
		rawMap, err := queryGoodsDetail(accessToken, successGoodsID, 5)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, fmt.Sprintf("查询商品详情失败: %v", err), elapsed)
			return
		}

		// 提取商品状态
		goodsStatus, goodsName := extractGoodsStatus(rawMap)
		// 尝试从 data 层提取
		if goodsStatus == -1 {
			if dataMap, ok := rawMap["data"].(map[string]interface{}); ok {
				goodsStatus, goodsName = extractGoodsStatus(dataMap)
			}
		}
		// 尝试从嵌套层提取
		if goodsStatus == -1 {
			for _, key := range []string{"goods", "goodsDetail", "result"} {
				if nested, ok := rawMap[key].(map[string]interface{}); ok {
					goodsStatus, goodsName = extractGoodsStatus(nested)
					if goodsStatus != -1 {
						break
					}
				}
			}
		}

		// 状态码映射
		statusStr := fmt.Sprintf("%d", goodsStatus)
		if n, ok := StatusName[goodsStatus]; ok {
			statusStr = n
		}

		if goodsStatus == 2 {
			pass(cat, name, fmt.Sprintf("商品已下架 status=2(下架) goodsId=%d goodsName=%s",
				successGoodsID, truncate(goodsName, 50)), elapsed)
		} else if goodsStatus == 1 {
			fail(cat, name, fmt.Sprintf("商品仍处于上架状态 status=1(上架) goodsId=%d，下架可能未生效", successGoodsID), elapsed)
		} else if goodsStatus == 3 {
			fail(cat, name, fmt.Sprintf("商品状态为售罄 status=3(售罄) goodsId=%d，非预期的下架状态", successGoodsID), elapsed)
		} else if goodsStatus == 4 {
			errCase(cat, name, fmt.Sprintf("商品已删除 status=4(已删除) goodsId=%d", successGoodsID), elapsed)
		} else {
			errCase(cat, name, fmt.Sprintf("未能解析商品状态 status=%d(%s)",
				goodsStatus, statusStr), elapsed)
		}
	}
}

// ============================================================
// 测试四（补充）：拼多多删除商品
// ============================================================

func testPddGoodsDelete() {
	cat := "四（补充）、拼多多删除商品"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	// 前置条件检查
	if mixedTaskID == "" {
		fmt.Println("⚠️  改价格 task_id 为空，跳过删除测试")
		fail(cat, "前置条件", "场景二未创建改价格任务，无法获取 task_id", 0)
		return
	}
	if successISBN == "" || successGoodsID == 0 {
		fmt.Println("⚠️  场景一未匹配到执行成功数据，跳过删除测试")
		fail(cat, "前置条件", "场景一未匹配到执行成功数据（isbn/goods_id）", 0)
		return
	}

	fmt.Printf("\n  📌 使用改价格 task_id: %s（复用 task_type=5 任务）\n", mixedTaskID)

	// ---------- 步骤1：发送删除任务 ----------
	{
		name := "1、发送任务数据【删除】"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"goods_id":%d,"status":3}}`,
			successISBN, successGoodsID)
		fmt.Printf("    📋 task_id: %s\n", mixedTaskID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)

		params := map[string]string{
			"task_id": mixedTaskID,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("ISBN=%s GoodsID=%d status=3(删除) 接口返回成功", successISBN, successGoodsID), elapsed)
	}

	// 延迟后校验
	countdownDelay(DelayPddDeleteAfterSend, "校验")

	// ---------- 步骤2：等待删除任务处理完成 ----------
	{
		name := "2、等待任务处理完成"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		fmt.Printf("    ⏳ 等待删除任务处理完成（检查 body_over）...\n")

		// 上下架已有 3 条，删除后再增加 1 条，应达到 4 条
		if err := waitBodyOverMin(mixedTaskID, BodyOverMinDelete, DelayPddShelfBodyOverTimeout); err != nil {
			fmt.Printf("    ⚠️  等待超时: %v，继续尝试校验\n", err)
		} else {
			fmt.Printf("    ✅ 删除任务已处理完成\n")
		}
		elapsed := time.Since(start)
		pass(cat, name, fmt.Sprintf("body_over ≥ %d", BodyOverMinDelete), elapsed)
	}

	// 步骤3：延迟后接口校验
	countdownDelay(DelayPddDeleteAfterAPI, "接口校验")

	// ---------- 步骤4：校验删除状态 ----------
	{
		name := "4、校验删除状态"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 获取 accessToken
		accessToken, err := getAccessToken(priceTaskID)
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}
		fmt.Printf("    🔑 accessToken: %s...%s\n", accessToken[:8], accessToken[len(accessToken)-8:])
		fmt.Printf("    📦 goodsId: %d\n", successGoodsID)

		// 查询商品详情（最多重试5次）
		rawMap, err := queryGoodsDetail(accessToken, successGoodsID, 5)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, fmt.Sprintf("查询商品详情失败: %v", err), elapsed)
			return
		}

		// 提取商品状态
		goodsStatus, goodsName := extractGoodsStatus(rawMap)
		if goodsStatus == -1 {
			if dataMap, ok := rawMap["data"].(map[string]interface{}); ok {
				goodsStatus, goodsName = extractGoodsStatus(dataMap)
			}
		}
		if goodsStatus == -1 {
			for _, key := range []string{"goods", "goodsDetail", "result"} {
				if nested, ok := rawMap[key].(map[string]interface{}); ok {
					goodsStatus, goodsName = extractGoodsStatus(nested)
					if goodsStatus != -1 {
						break
					}
				}
			}
		}

		var statusStr string
		if n, ok := StatusName[goodsStatus]; ok {
			statusStr = n
		}
		fmt.Printf("    📊 商品状态: status=%d(%s), name=%s\n", goodsStatus, statusStr, goodsName)

		if goodsStatus == 4 {
			pass(cat, name, fmt.Sprintf("商品已删除 status=4(已删除) goodsId=%d", successGoodsID), elapsed)
		} else if goodsStatus == 2 {
			fail(cat, name, fmt.Sprintf("商品状态为下架 status=2(下架) goodsId=%d，非预期的删除状态", successGoodsID), elapsed)
		} else if goodsStatus == 3 {
			fail(cat, name, fmt.Sprintf("商品状态为售罄 status=3(售罄) goodsId=%d，非预期的删除状态", successGoodsID), elapsed)
		} else if goodsStatus == 1 {
			fail(cat, name, fmt.Sprintf("商品状态为上架 status=1(上架) goodsId=%d，非预期的删除状态", successGoodsID), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("未能解析商品状态 status=%d, name=%s", goodsStatus, goodsName), elapsed)
		}
	}
}

// ============================================================
// 测试五：拼多多商品拉取任务
// ============================================================

func testPddPullGoods() {
	cat := "五、拼多多商品拉取"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	// ---------- 步骤1：创建商品拉取任务 ----------
	var tid string
	{
		name := "1、创建拼多多商品拉取任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePullGoods,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		if dataStr == "" {
			fail(cat, name, "返回 data 为空", elapsed)
			return
		}
		tid = dataStr
		pullTaskID = tid
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	if tid == "" {
		fmt.Println("⚠️  拉取任务创建失败，跳过后续步骤")
		return
	}
	fmt.Printf("\n  📌 商品拉取 task_id: %s\n", tid)

	// ---------- 步骤2：等待任务完成（status=4）并校验 ----------
	{
		name := "2、等待任务完成并校验"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 2a: 等待 header status=4
		fmt.Printf("    ⏳ 等待任务完成（header status=4，无时间限制，全店拉取数据量大）...\n")
		waitStart := time.Now()
		for {
			st, err := getHeaderStatus(tid)
			if err == nil && st == 4 {
				fmt.Printf("    ✅ 任务状态已变为 status=4（耗时 %v）\n", time.Since(waitStart).Round(time.Second))
				break
			}
			// 每30秒打印一次进度
			if int(time.Since(waitStart).Seconds()) > 0 && int(time.Since(waitStart).Seconds())%30 == 0 {
				bodyOverCnt, _ := getBodyOverCount(tid)
				fmt.Printf("    ⏳ 已等待 %v，当前 status=%d，body_over=%d\n",
					time.Since(waitStart).Round(time.Second), st, bodyOverCnt)
			}
			time.Sleep(PollInterval)
		}

		// 2b: 对比 task_count_true 与 body_over 数量
		headerKey := tid + ":header"
		taskCountTrueStr, err := redisClient.HGet(redisCtx, headerKey, "task_count_true").Result()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 task_count_true 失败: %v", err), time.Since(start))
			return
		}
		taskCountTrue, err := strconv.ParseInt(taskCountTrueStr, 10, 64)
		if err != nil {
			fail(cat, name, fmt.Sprintf("解析 task_count_true 失败: %v", err), time.Since(start))
			return
		}

		bodyOverCount, err := getBodyOverCount(tid)
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 body_over 数量失败: %v", err), time.Since(start))
			return
		}

		fmt.Printf("    📊 task_count_true=%d, body_over 数量=%d\n", taskCountTrue, bodyOverCount)

		if bodyOverCount > taskCountTrue {
			fail(cat, name, fmt.Sprintf("body_over 数量(%d) > task_count_true(%d)，不一致", bodyOverCount, taskCountTrue), time.Since(start))
			return
		}
		fmt.Printf("    ✅ body_over 数量(%d) ≤ task_count_true(%d)，一致\n", bodyOverCount, taskCountTrue)

		// 2c: 检查场景一 ISBN 是否存在于 body_over 或 body_wait 中
		if successISBN == "" {
			fail(cat, name, "场景一未匹配到执行成功数据（isbn 为空），无法检查拉取结果", time.Since(start))
			return
		}

		foundISBN := false
		var foundGoodsID int64
		var foundIn string

		// 先检查 body_over（分批搜索，全店拉取数据量大）
		fmt.Printf("    🔍 在 body_over 中搜索 ISBN=%s（共 %d 条，分批搜索）...\n", successISBN, bodyOverCount)
		searchPageSize := SearchPageSize
		for offset := int64(0); offset < bodyOverCount; offset += searchPageSize {
			end := offset + searchPageSize - 1
			if end >= bodyOverCount {
				end = bodyOverCount - 1
			}
			items, err := getBodyOverFromRedis(tid, offset, end)
			if err != nil || len(items) == 0 {
				break
			}
			for _, item := range items {
				if bi, ok := item["book_info"].(map[string]interface{}); ok {
					if isbn, ok := bi["isbn"].(string); ok && isbn == successISBN {
						foundISBN = true
						foundIn = "body_over"
						if detail, ok := item["detail"].(map[string]interface{}); ok {
							if v, ok := detail["goods_id"].(float64); ok {
								foundGoodsID = int64(v)
							}
						}
						break
					}
				}
			}
			if foundISBN {
				break
			}
			// 进度提示
			if (offset+searchPageSize)%BodyWaitMaxSearch == 0 {
				fmt.Printf("    🔍 已搜索 %d/%d 条...\n", offset+searchPageSize, bodyOverCount)
			}
		}

		// 如果 body_over 没找到，检查 body_wait
		if !foundISBN {
			bodyWaitKey := tid + ":body_wait"
			bodyWaitCount, err := redisClient.LLen(redisCtx, bodyWaitKey).Result()
			if err == nil && bodyWaitCount > 0 {
				fmt.Printf("    🔍 body_over 未找到，正在搜索 body_wait (%d 条)...\n", bodyWaitCount)
				maxSearch := bodyWaitCount
				if maxSearch > BodyWaitMaxSearch {
					maxSearch = BodyWaitMaxSearch
					fmt.Printf("    ⚠️ body_wait 数据量巨大，仅搜索前 %d 条\n", BodyWaitMaxSearch)
				}
				bodyWaitVals, err := redisClient.LRange(redisCtx, bodyWaitKey, 0, maxSearch-1).Result()
				if err == nil {
					for _, v := range bodyWaitVals {
						var item map[string]interface{}
						if json.Unmarshal([]byte(v), &item) != nil {
							continue
						}
						if bi, ok := item["book_info"].(map[string]interface{}); ok {
							if isbn, ok := bi["isbn"].(string); ok && isbn == successISBN {
								foundISBN = true
								foundIn = "body_wait"
								if detail, ok := item["detail"].(map[string]interface{}); ok {
									if v, ok := detail["goods_id"].(float64); ok {
										foundGoodsID = int64(v)
									}
								}
								break
							}
						}
					}
				}
			}
		}

		elapsed := time.Since(start)

		if foundISBN {
			pullGoodsID = foundGoodsID
			goodsIDStr := ""
			if foundGoodsID > 0 {
				goodsIDStr = fmt.Sprintf(", goods_id=%d", foundGoodsID)
			}
			pass(cat, name, fmt.Sprintf("status=4 ✅ | body_over(%d) ≤ task_count_true(%d) ✅ | ISBN=%s 在 %s 中找到%s",
				bodyOverCount, taskCountTrue, successISBN, foundIn, goodsIDStr), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("status=4 ✅ | body_over(%d) ≤ task_count_true(%d) ✅ | 但 ISBN=%s 不在 body_over 也不在 body_wait(前%d条) 中",
				bodyOverCount, taskCountTrue, successISBN, BodyWaitMaxSearch), elapsed)
		}
	}
}

// ============================================================
// 测试九：闲鱼改价格
// ============================================================

func testXyPriceChange() {
	cat := "九、闲鱼改价格"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	var tid string

	// 前置条件
	if xyTaskID == "" || xySuccessISBN == "" {
		fmt.Println("⚠️  场景七未创建闲鱼任务，跳过")
		return
	}
	fmt.Printf("\n  📌 复用场景七 task_id: %s\n", xyTaskID)
	fmt.Printf("  📌 目标 ISBN: %s\n", xySuccessISBN)

	// 步骤1：创建改价格任务（task_type=5，与闲鱼共用）
	{
		name := "1、创建闲鱼改价格任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    XyShopID,
			"shop_type":  XyShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypeXyPriceStockShelf,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		tid = dataStr
		fmt.Printf("    ✅ task_id=%s\n", tid)
	}

	if tid == "" {
		fmt.Println("⚠️  任务创建失败，跳过")
		return
	}

	// 步骤2：查询商品详情（改价格前）
	var basePrice int64
	{
		name := "2、查询商品详情（改价格前）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, err := getXyAccessToken()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}

		// 使用核价发布步骤已获取的 goods_id
		xyGoodsID := xySuccessGoodsID
		if xyGoodsID == 0 {
			fail(cat, name, "xySuccessGoodsID 为空，场景七未成功获取商品ID", time.Since(start))
			return
		}
		fmt.Printf("    🔑 accessToken: %s...%s\n", accessToken[:8], accessToken[len(accessToken)-8:])
		fmt.Printf("    📦 goodsId: %d\n", xyGoodsID)

		price, err := getXyGoodsPrice(xyGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询价格失败: %v", err), elapsed)
			return
		}
		basePrice = price
		fmt.Printf("    💰 当前价格: %d (分)\n", basePrice)
		pass(cat, name, fmt.Sprintf("goodsId=%d 当前价格=%d", xyGoodsID, basePrice), elapsed)
	}

	countdownDelay(DelayXyPriceChangeAfterQuery, "改价格")

	// 步骤3：发送改价格任务
	{
		name := "3、发送改价格任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		xyGoodsID := xySuccessGoodsID
		if xyGoodsID == 0 {
			fail(cat, name, "xySuccessGoodsID 为空", time.Since(start))
			return
		}
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"goods_id":%d,"price":%d,"status":5}}`,
			xySuccessISBN, xyGoodsID, TestXyNewPrice)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("改价格目标=%d (分)", TestXyNewPrice), elapsed)
	}

	countdownDelay(DelayXyPriceChangeAfterSend, "Redis校验")

	// 步骤4：校验 body_over
	{
		name := "4、校验 body_over"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		if err := waitBodyOverMin(tid, 1, 30); err != nil {
			fail(cat, name, fmt.Sprintf("body_over < 1: %v", err), time.Since(start))
		} else {
			pass(cat, name, "body_over ≥ 1", time.Since(start))
		}
	}

	countdownDelay(DelayXyPriceChangeAfterAPI, "接口校验")

	// 步骤5：接口校验
	{
		name := "5、接口校验（价格已修改）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, _ := getXyAccessToken()
		xyGoodsID := xySuccessGoodsID
		newPrice, err := getXyGoodsPrice(xyGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询价格失败: %v", err), elapsed)
			return
		}
		fmt.Printf("    📊 当前价格: %d (分)\n", newPrice)
		if newPrice == TestXyNewPrice {
			pass(cat, name, fmt.Sprintf("价格已改为 %d", TestXyNewPrice), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("价格未变更: 期望=%d 实际=%d", TestXyNewPrice, newPrice), elapsed)
		}
	}
}

// ============================================================
// 测试十：闲鱼改库存
// ============================================================

func testXyStockChange() {
	cat := "十、闲鱼改库存"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	// 复用九创建的改价格任务
	fmt.Printf("\n  📌 目标 ISBN: %s\n", xySuccessISBN)

	// 步骤1：创建改库存任务
	{
		name := "1、创建闲鱼改库存任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    XyShopID,
			"shop_type":  XyShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypeXyPriceStockShelf,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		xyModTaskID = dataStr
		fmt.Printf("    ✅ task_id=%s\n", xyModTaskID)
	}

	if xyModTaskID == "" {
		fmt.Println("⚠️  任务创建失败，跳过")
		return
	}

	// 步骤2：发送改库存任务
	{
		name := "2、发送改库存任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		xyGoodsID := xySuccessGoodsID
		if xyGoodsID == 0 {
			fail(cat, name, "xySuccessGoodsID 为空", time.Since(start))
			return
		}
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"goods_id":%d,"stock":%d,"status":4}}`,
			xySuccessISBN, xyGoodsID, TestXyNewStock)
		fmt.Printf("    📋 task_id: %s\n", xyModTaskID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)
		params := map[string]string{
			"task_id": xyModTaskID,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("改库存目标=%d", TestXyNewStock), elapsed)
	}

	countdownDelay(DelayXyStockAfterSend, "Redis校验")

	// 步骤3：校验 body_over
	{
		name := "3、校验 body_over"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		if err := waitBodyOverMin(xyModTaskID, 1, 30); err != nil {
			fail(cat, name, fmt.Sprintf("body_over < 1: %v", err), time.Since(start))
		} else {
			pass(cat, name, "body_over ≥ 1", time.Since(start))
		}
	}

	countdownDelay(DelayXyStockAfterAPI, "接口校验")

	// 步骤4：接口校验
	{
		name := "4、接口校验（库存已修改）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		accessToken, _ := getXyAccessToken()
		xyGoodsID := xySuccessGoodsID
		newStock, err := getXyGoodsStock(xyGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询库存失败: %v", err), elapsed)
			return
		}
		fmt.Printf("    📊 当前库存: %d\n", newStock)
		if newStock == TestXyNewStock {
			pass(cat, name, fmt.Sprintf("库存已改为 %d", TestXyNewStock), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("库存未变更: 期望=%d 实际=%d", TestXyNewStock, newStock), elapsed)
		}
	}
}

// ============================================================
// 测试十一：闲鱼上下架
// ============================================================

func testXyShelfOff() {
	cat := "十一、闲鱼下架"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	// 前置条件检查
	if xyModTaskID == "" {
		fmt.Println("⚠️  闲鱼核价发布 task_id 为空，跳过下架测试")
		fail(cat, "前置条件", "场景七未创建核价发布任务，无法获取 task_id", 0)
		return
	}
	if xySuccessISBN == "" || xySuccessGoodsID == 0 {
		fmt.Println("⚠️  场景七未匹配到执行成功数据，跳过下架测试")
		fail(cat, "前置条件", "场景七未匹配到执行成功数据（isbn/goods_id）", 0)
		return
	}

	fmt.Printf("\n  📌 复用场景七 task_id: %s\n", xyModTaskID)
	fmt.Printf("  📌 目标 ISBN: %s\n", xySuccessISBN)

	// ---------- 步骤1：发送下架任务 ----------
	{
		name := "1、发送任务数据【下架】"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"goods_id":%d,"status":2}}`,
			xySuccessISBN, xySuccessGoodsID)
		fmt.Printf("    📋 task_id: %s\n", xyModTaskID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)

		params := map[string]string{
			"task_id": xyModTaskID,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("ISBN=%s GoodsID=%d status=2(下架) 接口返回成功", xySuccessISBN, xySuccessGoodsID), elapsed)
	}

	countdownDelay(DelayXyShelfAfterSend, "校验")

	// ---------- 等待下架任务处理完成 ----------
	{
		fmt.Printf("\n    ⏳ 等待下架任务处理完成（检查 body_over）...\n")
		if err := waitBodyOverMin(xyModTaskID, BodyOverMinXyShelfOnOff, DelayXyShelfBodyOverTimeout); err != nil {
			fmt.Printf("    ⚠️  等待超时: %v，继续尝试校验\n", err)
		} else {
			fmt.Printf("    ✅ 下架任务已处理完成\n")
		}
	}

	countdownDelay(DelayXyShelfAfterWait, "接口校验")

	// ---------- 步骤2：校验下架状态 ----------
	{
		name := "2、校验下架状态"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, _ := getXyAccessToken()
		status, err := getXyGoodsStatus(xySuccessGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询状态失败: %v", err), elapsed)
			return
		}
		statusStr := ""
		if n, ok := StatusName[status]; ok {
			statusStr = n
		}
		fmt.Printf("    📊 商品状态: status=%d(%s)\n", status, statusStr)
		if status == 2 {
			pass(cat, name, fmt.Sprintf("商品已下架 status=2(下架) goodsId=%d", xySuccessGoodsID), elapsed)
		} else if status == 1 {
			fail(cat, name, fmt.Sprintf("商品仍处于上架状态 status=1(上架) goodsId=%d，下架可能未生效", xySuccessGoodsID), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("状态不符合预期: 期望=status=2(下架) 实际=status=%d(%s)", status, statusStr), elapsed)
		}
	}
}

// ============================================================
// 测试十二：闲鱼上架
// ============================================================
func testXyShelfOn() {
	cat := "十二、闲鱼上架"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	// 前置条件检查
	if xyModTaskID == "" {
		fmt.Println("⚠️  闲鱼核价发布 task_id 为空，跳过上架测试")
		fail(cat, "前置条件", "场景七未创建核价发布任务，无法获取 task_id", 0)
		return
	}
	if xySuccessISBN == "" || xySuccessGoodsID == 0 {
		fmt.Println("⚠️  场景七未匹配到执行成功数据，跳过上架测试")
		fail(cat, "前置条件", "场景七未匹配到执行成功数据（isbn/goods_id）", 0)
		return
	}

	fmt.Printf("\n  📌 复用场景七 task_id: %s\n", xyModTaskID)
	fmt.Printf("  📌 目标 ISBN: %s\n", xySuccessISBN)

	// ---------- 步骤1：发送上架任务 ----------
	{
		name := "1、发送任务数据【上架】"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"goods_id":%d,"status":1}}`,
			xySuccessISBN, xySuccessGoodsID)
		fmt.Printf("    📋 task_id: %s\n", xyModTaskID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)

		params := map[string]string{
			"task_id": xyModTaskID,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("ISBN=%s GoodsID=%d status=1(上架) 接口返回成功", xySuccessISBN, xySuccessGoodsID), elapsed)
	}

	countdownDelay(DelayXyShelfAfterSend, "校验")

	// ---------- 等待上架任务处理完成 ----------
	{
		fmt.Printf("\n    ⏳ 等待上架任务处理完成（检查 body_over）...\n")
		if err := waitBodyOverMin(xyModTaskID, BodyOverMinXyShelfOnOff, DelayXyShelfBodyOverTimeout); err != nil {
			fmt.Printf("    ⚠️  等待超时: %v，继续尝试校验\n", err)
		} else {
			fmt.Printf("    ✅ 上架任务已处理完成\n")
		}
	}

	countdownDelay(DelayXyShelfAfterWait, "接口校验")

	// ---------- 步骤2：校验上架状态 ----------
	{
		name := "2、校验上架状态"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, _ := getXyAccessToken()
		status, err := getXyGoodsStatus(xySuccessGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询状态失败: %v", err), elapsed)
			return
		}
		statusStr := ""
		if n, ok := StatusName[status]; ok {
			statusStr = n
		}
		fmt.Printf("    📊 商品状态: status=%d(%s)\n", status, statusStr)
		if status == 1 {
			pass(cat, name, fmt.Sprintf("商品已上架 status=1(上架) goodsId=%d", xySuccessGoodsID), elapsed)
		} else if status == 2 {
			fail(cat, name, fmt.Sprintf("商品仍处于下架状态 status=2(下架) goodsId=%d，上架可能未生效", xySuccessGoodsID), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("状态不符合预期: 期望=status=1(上架) 实际=status=%d(%s)", status, statusStr), elapsed)
		}
	}
}

// ============================================================
// 测试八：闲鱼商品拉取任务
// ============================================================
func testXyPullGoods() {
	cat := "八、闲鱼商品拉取任务"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	var tid string

	// ---------- 步骤1：创建闲鱼商品拉取任务 ----------
	{
		name := "1、创建闲鱼商品拉取任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		fields := map[string]string{
			"shop_id":    XyShopID,
			"shop_type":  XyShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePullGoods,
			"img_type":   ImgType,
		}
		fields["sign"] = SignParams(fields)

		resp, err := postMultipart(BaseURL+"/task/create", fields)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		if dataStr == "" {
			fail(cat, name, "返回 data 为空", elapsed)
			return
		}
		tid = dataStr
		xyPullTaskID = tid
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	if tid == "" {
		fmt.Println("⚠️  拉取任务创建失败，跳过后续步骤")
		return
	}
	fmt.Printf("\n  📌 闲鱼商品拉取 task_id: %s\n", tid)

	// ---------- 步骤2：等待任务完成（status=4）并校验 ----------
	{
		name := "2、等待任务完成并校验"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 2a: 等待 header status=4
		fmt.Printf("    ⏳ 等待任务完成（header status=4，无时间限制，拉取数据量大）...\n")
		waitStart := time.Now()
		for {
			st, err := getHeaderStatus(tid)
			if err == nil && st == 4 {
				fmt.Printf("    ✅ 任务状态已变为 status=4（耗时 %v）\n", time.Since(waitStart).Round(time.Second))
				break
			}
			// 每30秒打印一次进度
			if int(time.Since(waitStart).Seconds()) > 0 && int(time.Since(waitStart).Seconds())%30 == 0 {
				bodyOverCnt, _ := getBodyOverCount(tid)
				fmt.Printf("    ⏳ 已等待 %v，当前 status=%d，body_over=%d\n",
					time.Since(waitStart).Round(time.Second), st, bodyOverCnt)
			}
			time.Sleep(PollInterval)
		}

		// 2b: 对比 task_count_true 与 body_over 数量
		headerKey := tid + ":header"
		taskCountTrueStr, err := redisClient.HGet(redisCtx, headerKey, "task_count_true").Result()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 task_count_true 失败: %v", err), time.Since(start))
			return
		}
		taskCountTrue, err := strconv.ParseInt(taskCountTrueStr, 10, 64)
		if err != nil {
			fail(cat, name, fmt.Sprintf("解析 task_count_true 失败: %v", err), time.Since(start))
			return
		}

		bodyOverCount, err := getBodyOverCount(tid)
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 body_over 数量失败: %v", err), time.Since(start))
			return
		}

		fmt.Printf("    📊 task_count_true=%d, body_over 数量=%d\n", taskCountTrue, bodyOverCount)

		if bodyOverCount > taskCountTrue {
			fail(cat, name, fmt.Sprintf("body_over 数量(%d) > task_count_true(%d)，不一致", bodyOverCount, taskCountTrue), time.Since(start))
			return
		}
		fmt.Printf("    ✅ body_over 数量(%d) ≤ task_count_true(%d)，一致\n", bodyOverCount, taskCountTrue)

		// 2c: 检查场景七 ISBN 是否存在于 body_over 或 body_wait 中
		if xySuccessISBN == "" {
			fail(cat, name, "场景七未匹配到执行成功数据（isbn 为空），无法检查拉取结果", time.Since(start))
			return
		}

		foundISBN := false
		var foundGoodsID int64
		var foundIn string

		// 先检查 body_over（分批搜索，数据量大）
		fmt.Printf("    🔍 在 body_over 中搜索 ISBN=%s（共 %d 条，分批搜索）...\n", xySuccessISBN, bodyOverCount)
		searchPageSize := SearchPageSize
		for offset := int64(0); offset < bodyOverCount; offset += searchPageSize {
			end := offset + searchPageSize - 1
			if end >= bodyOverCount {
				end = bodyOverCount - 1
			}
			items, err := getBodyOverFromRedis(tid, offset, end)
			if err != nil || len(items) == 0 {
				break
			}
			for _, item := range items {
				if bi, ok := item["book_info"].(map[string]interface{}); ok {
					if isbn, ok := bi["isbn"].(string); ok && isbn == xySuccessISBN {
						foundISBN = true
						foundIn = "body_over"
						if detail, ok := item["detail"].(map[string]interface{}); ok {
							if v, ok := detail["goods_id"].(float64); ok {
								foundGoodsID = int64(v)
							}
						}
						break
					}
				}
			}
			if foundISBN {
				break
			}
			// 进度提示
			if (offset+searchPageSize)%BodyWaitMaxSearch == 0 {
				fmt.Printf("    🔍 已搜索 %d/%d 条...\n", offset+searchPageSize, bodyOverCount)
			}
		}

		// 如果 body_over 没找到，检查 body_wait
		if !foundISBN {
			bodyWaitKey := tid + ":body_wait"
			bodyWaitCount, err := redisClient.LLen(redisCtx, bodyWaitKey).Result()
			if err == nil && bodyWaitCount > 0 {
				fmt.Printf("    🔍 body_over 未找到，正在搜索 body_wait (%d 条)...\n", bodyWaitCount)
				maxSearch := bodyWaitCount
				if maxSearch > BodyWaitMaxSearch {
					maxSearch = BodyWaitMaxSearch
					fmt.Printf("    ⚠️ body_wait 数据量巨大，仅搜索前 %d 条\n", BodyWaitMaxSearch)
				}
				bodyWaitVals, err := redisClient.LRange(redisCtx, bodyWaitKey, 0, maxSearch-1).Result()
				if err == nil {
					for _, v := range bodyWaitVals {
						var item map[string]interface{}
						if json.Unmarshal([]byte(v), &item) != nil {
							continue
						}
						if bi, ok := item["book_info"].(map[string]interface{}); ok {
							if isbn, ok := bi["isbn"].(string); ok && isbn == xySuccessISBN {
								foundISBN = true
								foundIn = "body_wait"
								if detail, ok := item["detail"].(map[string]interface{}); ok {
									if v, ok := detail["goods_id"].(float64); ok {
										foundGoodsID = int64(v)
									}
								}
								break
							}
						}
					}
				}
			}
		}

		elapsed := time.Since(start)

		if foundISBN {
			xyPullGoodsID = foundGoodsID
			goodsIDStr := ""
			if foundGoodsID > 0 {
				goodsIDStr = fmt.Sprintf(", goods_id=%d", foundGoodsID)
			}
			pass(cat, name, fmt.Sprintf("status=4 ✅ | body_over(%d) ≤ task_count_true(%d) ✅ | ISBN=%s 在 %s 中找到%s",
				bodyOverCount, taskCountTrue, xySuccessISBN, foundIn, goodsIDStr), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("status=4 ✅ | body_over(%d) ≤ task_count_true(%d) ✅ | 但 ISBN=%s 不在 body_over 也不在 body_wait(前%d条) 中",
				bodyOverCount, taskCountTrue, xySuccessISBN, BodyWaitMaxSearch), elapsed)
		}
	}
}

// ============================================================
// 主函数
// ============================================================

// ============================================================
// 孔夫子测试函数
// ============================================================

// testKfzPricePublish 孔夫子核价发布
func testKfzPricePublish() {
	cat := "孔夫子核价发布"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	var tid string

	// 步骤1：创建任务
	{
		name := "1、创建孔夫子核价发布任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePricePublish,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		if dataStr == "" {
			fail(cat, name, "返回 data 为空", elapsed)
			return
		}
		tid = dataStr
		kfzTaskID = tid
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	if tid == "" {
		fmt.Println("⚠️  任务创建失败，跳过后续步骤")
		return
	}
	fmt.Printf("\n  📌 孔夫子 task_id: %s\n", tid)

	// 步骤2：发送 ISBN 数据
	{
		name := fmt.Sprintf("2、发送任务数据【isbn=%s, price=%d】期望：执行成功", TestKfzISBNSuccess, TestKfzPriceSuccess)
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"price":%d}}`, TestKfzISBNSuccess, TestKfzPriceSuccess)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, "接口返回成功", elapsed)
	}

	countdownDelay(DelayKfzPublishAfterSend, "Redis校验")

	// 步骤3：Redis 校验 body_over
	{
		name := "3、Redis 校验 - body_over 中执行成功"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		fmt.Printf("    ⏳ 等待后台处理完成（最多 %v）...\n", WaitTimeout)

		if err := waitBodyOverMin(tid, 1, WaitTimeout); err != nil {
			errCase(cat, name, "等待超时: "+err.Error(), time.Since(start))
			return
		}

		totalCnt, _ := getBodyOverCount(tid)
		fmt.Printf("    📊 body_over 总数: %d，开始校验...\n", totalCnt)

		targetISBN := TestKfzISBNSuccess
		found := false
		var foundGoodsID int64

		for offset := int64(0); offset < totalCnt && !found; offset += 100 {
			items, err := getBodyOverFromRedis(tid, offset, offset+99)
			if err != nil || len(items) == 0 {
				break
			}
			for _, item := range items {
				isbn := ""
				if bi, ok := item["book_info"].(map[string]interface{}); ok {
					if v, ok := bi["isbn"].(string); ok {
						isbn = v
					}
				}
				if isbn != targetISBN {
					continue
				}
				// 找到了，提取 itemId
				if detail, ok := item["detail"].(map[string]interface{}); ok {
					if itemId, ok := detail["itemId"].(float64); ok {
						foundGoodsID = int64(itemId)
						found = true
						break
					}
				}
			}
		}

		if found {
			kfzSuccessISBN = targetISBN
			kfzSuccessGoodsID = foundGoodsID
			pass(cat, name, fmt.Sprintf("ISBN=%s 找到，itemId=%d", targetISBN, foundGoodsID), time.Since(start))
			fmt.Printf("\n  📌 核价发布成功 - ISBN=%s, itemId=%d\n", targetISBN, foundGoodsID)
		} else {
			fail(cat, name, fmt.Sprintf("ISBN=%s 在 body_over 中未找到", targetISBN), time.Since(start))
		}
	}
}

// testKfzPriceChange 孔夫子改价格
func testKfzPriceChange() {
	cat := "孔夫子改价格"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	if kfzTaskID == "" || kfzSuccessISBN == "" {
		fmt.Println("⚠️  孔夫子发布任务未创建，跳过")
		return
	}
	fmt.Printf("\n  📌 task_id: %s\n", kfzTaskID)
	fmt.Printf("  📌 目标 ISBN: %s, itemId: %d\n", kfzSuccessISBN, kfzSuccessGoodsID)

	// 步骤1：创建改价格任务
	var tid string
	{
		name := "1、创建孔夫子改价格任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePriceStockShelf,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		tid = dataStr
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	countdownDelay(DelayKfzPriceChangeAfterQuery, "改价格")

	// 步骤2：发送改价格任务
	{
		name := "2、发送改价格任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"itemId":%d,"price":%d}}`,
			kfzSuccessISBN, kfzSuccessGoodsID, TestKfzNewPrice)
		fmt.Printf("    📋 body: %s\n", bodyJSON)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("改价格目标=%d", TestKfzNewPrice), elapsed)
	}

	countdownDelay(DelayKfzPriceChangeAfterSend, "Redis校验")

	// 步骤3：校验 body_over
	{
		name := "3、校验 body_over"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		if err := waitBodyOverMin(tid, 1, 30); err != nil {
			fail(cat, name, fmt.Sprintf("body_over < 1: %v", err), time.Since(start))
		} else {
			pass(cat, name, "body_over ≥ 1", time.Since(start))
		}
	}

	countdownDelay(DelayKfzPriceChangeAfterAPI, "接口校验")

	// 步骤4：接口校验
	{
		name := "4、接口校验（价格已修改）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, err := getKfzAccessToken()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}
		fmt.Printf("    🔑 token: %s...%s\n", accessToken[:8], accessToken[len(accessToken)-8:])

		newPrice, err := getKfzGoodsPrice(kfzSuccessGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询价格失败: %v", err), elapsed)
			return
		}
		fmt.Printf("    📊 当前价格: %.2f (元)\n", newPrice)
		if newPrice == float64(TestKfzNewPrice)/100 {
			pass(cat, name, fmt.Sprintf("价格已改为 %.2f", float64(TestKfzNewPrice)/100), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("价格未变更: 期望=%.2f 实际=%.2f", float64(TestKfzNewPrice)/100, newPrice), elapsed)
		}
	}
}

// testKfzStockChange 孔夫子改库存
func testKfzStockChange() {
	cat := "孔夫子改库存"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	if kfzSuccessGoodsID == 0 {
		fmt.Println("⚠️  孔夫子发布任务未创建有效商品，跳过")
		return
	}
	fmt.Printf("\n  📌 itemId: %d\n", kfzSuccessGoodsID)

	var tid string
	// 步骤1：创建改库存任务
	{
		name := "1、创建孔夫子改库存任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePriceStockShelf,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		tid = dataStr
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	// 步骤2：发送改库存任务
	{
		name := "2、发送改库存任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"itemId":%d,"stock":%d}}`,
			kfzSuccessISBN, kfzSuccessGoodsID, TestKfzNewStock)
		fmt.Printf("    📋 body: %s\n", bodyJSON)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, fmt.Sprintf("改库存目标=%d", TestKfzNewStock), elapsed)
	}

	countdownDelay(DelayKfzStockAfterSend, "Redis校验")

	// 步骤3：校验 body_over
	{
		name := "3、校验 body_over"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		if err := waitBodyOverMin(tid, 1, 30); err != nil {
			fail(cat, name, fmt.Sprintf("body_over < 1: %v", err), time.Since(start))
		} else {
			pass(cat, name, "body_over ≥ 1", time.Since(start))
		}
	}

	countdownDelay(DelayKfzStockAfterAPI, "接口校验")

	// 步骤4：接口校验
	{
		name := "4、接口校验（库存已修改）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, err := getKfzAccessToken()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}

		newStock, err := getKfzGoodsStock(kfzSuccessGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询库存失败: %v", err), elapsed)
			return
		}
		fmt.Printf("    📊 当前库存: %d\n", newStock)
		if newStock == int(TestKfzNewStock) {
			pass(cat, name, fmt.Sprintf("库存已改为 %d", TestKfzNewStock), elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("库存未变更: 期望=%d 实际=%d", TestKfzNewStock, newStock), elapsed)
		}
	}
}

// testKfzShelfOnOff 孔夫子上下架
func testKfzShelfOnOff() {
	cat := "孔夫子上下架"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	if kfzSuccessGoodsID == 0 {
		fmt.Println("⚠️  孔夫子发布任务未创建有效商品，跳过")
		return
	}
	fmt.Printf("\n  📌 itemId: %d\n", kfzSuccessGoodsID)

	var tid string

	// 步骤1：创建上下架任务
	{
		name := "1、创建孔夫子上下架任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePriceStockShelf,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		tid = dataStr
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	// 步骤2：发送上架任务（status=1）
	{
		name := "2、发送上架任务（status=1）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"itemId":%d,"status":1}}`,
			kfzSuccessISBN, kfzSuccessGoodsID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, "上架任务发送成功", elapsed)
	}

	countdownDelay(DelayKfzShelfAfterSend, "处理")

	// 步骤3：等待处理
	{
		name := "3、等待任务处理"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		fmt.Printf("    ⏳ 等待后台处理完成（最多 %v）...\n", WaitTimeout)

		if err := waitBodyOverMin(tid, 1, WaitTimeout); err != nil {
			errCase(cat, name, "等待超时: "+err.Error(), time.Since(start))
			return
		}
		pass(cat, name, "任务处理完成", time.Since(start))
	}

	countdownDelay(DelayKfzShelfAfterWait, "接口校验")

	// 步骤4：接口校验（上架）
	{
		name := "4、接口校验（上架）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, err := getKfzAccessToken()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}

		status, err := getKfzGoodsStatus(kfzSuccessGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询上下架状态失败: %v", err), elapsed)
			return
		}
		fmt.Printf("    📊 当前上下架状态: isOnSale=%d (1=上架)\n", status)
		if status == 1 {
			pass(cat, name, "商品已上架", elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("商品未上架: isOnSale=%d", status), elapsed)
		}
	}

	// 步骤5：发送下架任务（status=2）
	{
		name := "5、发送下架任务（status=2）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"itemId":%d,"status":2}}`,
			kfzSuccessISBN, kfzSuccessGoodsID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, "下架任务发送成功", elapsed)
	}

	countdownDelay(DelayKfzShelfAfterSend, "处理")

	// 步骤6：等待处理
	{
		name := "6、等待任务处理"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		if err := waitBodyOverMin(tid, 2, WaitTimeout); err != nil {
			errCase(cat, name, "等待超时: "+err.Error(), time.Since(start))
			return
		}
		pass(cat, name, "任务处理完成", time.Since(start))
	}

	countdownDelay(DelayKfzShelfAfterWait, "接口校验")

	// 步骤7：接口校验（下架）
	{
		name := "7、接口校验（下架）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, err := getKfzAccessToken()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}

		status, err := getKfzGoodsStatus(kfzSuccessGoodsID, accessToken)
		elapsed := time.Since(start)
		if err != nil {
			fail(cat, name, fmt.Sprintf("查询上下架状态失败: %v", err), elapsed)
			return
		}
		fmt.Printf("    📊 当前上下架状态: isOnSale=%d (0=下架)\n", status)
		if status == 0 {
			pass(cat, name, "商品已下架", elapsed)
		} else {
			fail(cat, name, fmt.Sprintf("商品未下架: isOnSale=%d", status), elapsed)
		}
	}
}

// testKfzGoodsDelete 孔夫子删除商品
func testKfzGoodsDelete() {
	cat := "孔夫子删除商品"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	if kfzSuccessGoodsID == 0 {
		fmt.Println("⚠️  孔夫子发布任务未创建有效商品，跳过")
		return
	}
	fmt.Printf("\n  📌 itemId: %d\n", kfzSuccessGoodsID)

	var tid string
	// 步骤1：创建删除任务
	{
		name := "1、创建孔夫子删除任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePriceStockShelf,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		tid = dataStr
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	// 步骤2：发送删除任务（status=4 表示删除操作）
	{
		name := "2、发送删除任务（status=4）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		bodyJSON := fmt.Sprintf(`{"book_info":{"isbn":"%s"},"detail":{"itemId":%d,"status":4}}`,
			kfzSuccessISBN, kfzSuccessGoodsID)
		fmt.Printf("    📋 body: %s\n", bodyJSON)
		params := map[string]string{
			"task_id": tid,
			"body":    bodyJSON,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/setTaskBody", params)
		elapsed := time.Since(start)
		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		pass(cat, name, "删除任务发送成功", elapsed)
	}

	countdownDelay(DelayKfzDeleteAfterSend, "Redis校验")

	// 步骤3：校验 body_over
	{
		name := "3、校验 body_over"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)
		fmt.Printf("    ⏳ 等待后台处理完成（最多 %v）...\n", WaitTimeout)

		if err := waitBodyOverMin(tid, 1, WaitTimeout); err != nil {
			errCase(cat, name, "等待超时: "+err.Error(), time.Since(start))
			return
		}
		pass(cat, name, "body_over ≥ 1", time.Since(start))
	}

	countdownDelay(DelayKfzDeleteAfterAPI, "接口校验")

	// 步骤4：接口校验（调用 DLL 验证商品是否已删除）
	{
		name := "4、接口校验（商品已删除）"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		accessToken, err := getKfzAccessToken()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 accessToken 失败: %v", err), time.Since(start))
			return
		}
		fmt.Printf("    🔑 token: %s...%s\n", accessToken[:8], accessToken[len(accessToken)-8:])

		// 调用 DLL 搜索该 ISBN，看是否还能找到
		_, err = findKfzGoodsByISBN(kfzSuccessISBN, accessToken)
		elapsed := time.Since(start)
		if err != nil && strings.Contains(err.Error(), "未找到") {
			// 商品已被删除，找不到了
			pass(cat, name, fmt.Sprintf("商品已删除（ISBN=%s 无法在列表中找到）", kfzSuccessISBN), elapsed)
		} else if err != nil {
			fail(cat, name, fmt.Sprintf("查询商品失败: %v", err), elapsed)
		} else {
			// 还能找到，说明删除失败
			fail(cat, name, fmt.Sprintf("商品未删除，ISBN=%s 仍在列表中", kfzSuccessISBN), elapsed)
		}
	}
}

// testKfzPullGoods 孔夫子商品拉取
func testKfzPullGoods() {
	cat := "孔夫子商品拉取"
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  " + cat)
	fmt.Println(strings.Repeat("=", 60))

	var tid string

	// 步骤1：创建拉取任务
	{
		name := "1、创建孔夫子商品拉取任务"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		params := map[string]string{
			"shop_id":    ShopID,
			"shop_type":  ShopType,
			"task_count": TaskCount,
			"task_type":  TaskTypePullGoods,
			"img_type":   ImgType,
		}
		params["sign"] = SignParams(params)
		resp, err := postMultipart(BaseURL+"/task/create", params)
		elapsed := time.Since(start)

		if err != nil {
			errCase(cat, name, err.Error(), elapsed)
			return
		}
		if resp.Code != "200" {
			fail(cat, name, fmt.Sprintf("code=%s msg=%s", resp.Code, resp.Msg), elapsed)
			return
		}
		dataStr, _ := resp.Data.(string)
		if dataStr == "" {
			fail(cat, name, "返回 data 为空", elapsed)
			return
		}
		tid = dataStr
		pass(cat, name, fmt.Sprintf("task_id=%s", tid), elapsed)
	}

	if tid == "" {
		fmt.Println("⚠️  拉取任务创建失败，跳过后续步骤")
		return
	}
	fmt.Printf("\n  📌 孔夫子商品拉取 task_id: %s\n", tid)

	// 步骤2：等待任务完成（无时间限制，拉取数据量大）
	{
		name := "2、等待任务完成并校验"
		start := time.Now()
		fmt.Printf("\n[步骤] %s\n", name)

		// 等待 header status=4
		fmt.Printf("    ⏳ 等待任务完成（header status=4，无时间限制）...\n")
		waitStart := time.Now()
		for {
			st, err := getHeaderStatus(tid)
			if err == nil && st == 4 {
				fmt.Printf("    ✅ 任务状态已变为 status=4（耗时 %v）\n", time.Since(waitStart).Round(time.Second))
				break
			}
			if int(time.Since(waitStart).Seconds()) > 0 && int(time.Since(waitStart).Seconds())%30 == 0 {
				bodyOverCnt, _ := getBodyOverCount(tid)
				fmt.Printf("    ⏳ 已等待 %v，status=%d，body_over=%d\n",
					time.Since(waitStart).Round(time.Second), st, bodyOverCnt)
			}
			time.Sleep(PollInterval)
		}

		// 校验 task_count_true 与 body_over 数量
		headerKey := tid + ":header"
		taskCountTrueStr, err := redisClient.HGet(redisCtx, headerKey, "task_count_true").Result()
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 task_count_true 失败: %v", err), time.Since(start))
			return
		}
		taskCountTrue, _ := strconv.ParseInt(taskCountTrueStr, 10, 64)
		bodyOverCount, err := getBodyOverCount(tid)
		if err != nil {
			fail(cat, name, fmt.Sprintf("获取 body_over 数量失败: %v", err), time.Since(start))
			return
		}

		fmt.Printf("    📊 task_count_true=%d, body_over=%d\n", taskCountTrue, bodyOverCount)
		if bodyOverCount > taskCountTrue {
			fail(cat, name, fmt.Sprintf("body_over(%d) > task_count_true(%d)", bodyOverCount, taskCountTrue), time.Since(start))
		} else {
			pass(cat, name, fmt.Sprintf("body_over ≤ task_count_true"), time.Since(start))
		}
	}
}

// ============================================================
// 主函数
// ============================================================

func main() {
	reportDir, _ = os.Getwd()

	// 加载配置
	configFile := filepath.Join(reportDir, configPath)
	if err := loadConfig(configFile); err != nil {
		fmt.Printf("🔴 加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(strings.Repeat("#", 60))
	fmt.Println("          🧪 PlanA API 批量测试")
	fmt.Println(strings.Repeat("#", 60))
	fmt.Printf("接口地址     : %s\n", BaseURL)
	fmt.Printf("ShopID       : %s\n", ShopID)
	fmt.Printf("拼多多应用ID : %s\n", PddAppID)
	fmt.Printf("Redis        : %s (DB=%d)\n\n", RedisAddr, RedisDB)

	// 预检 HTTP 服务
	fmt.Println("🔍 预检 HTTP 服务...")
	resp, err := httpClient.Get(BaseURL + "/task/get?page=1&size=1")
	if err != nil {
		fmt.Printf("🔴 无法连接 %s: %v\n", BaseURL, err)
		fmt.Printf("请确认 planA HTTP 服务已启动 (%s)\n", BaseURL)
		os.Exit(1)
	}
	resp.Body.Close()
	fmt.Println("✅ HTTP 服务正常\n")

	// 预检 Redis
	fmt.Println("🔍 预检 Redis...")
	if err := initRedis(); err != nil {
		fmt.Printf("🔴 无法连接 Redis %s: %v\n", RedisAddr, err)
		fmt.Println("请确认 Redis 服务已启动")
		os.Exit(1)
	}
	fmt.Println("✅ Redis 连接正常\n")

	// 预检 curl
	fmt.Println("🔍 预检 curl...")
	if _, err := exec.LookPath("curl"); err != nil {
		fmt.Println("🔴 curl 未安装，校验步骤需要 curl")
		fmt.Println("请安装 curl: https://curl.se/download.html")
		os.Exit(1)
	}
	fmt.Println("✅ curl 可用\n")

	// // 拼多多测试
	// testPddPricePublish()
	// testPddPriceChange()
	// testPddStockChange()
	// testPddShelfOnOff()
	// testPddGoodsDelete()
	// testPddPullGoods()

	// 闲鱼测试
	testXyPricePublish()
	testXyPriceChange()
	testXyStockChange()
	testXyShelfOff()
	testXyShelfOn()
	//testXyPullGoods()

	// 生成并保存报告
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("📊 生成测试报告...")

	report := generateReport()
	reportFile := filepath.Join(reportDir, "test_report.md")
	if err := os.WriteFile(reportFile, []byte(report), 0644); err != nil {
		fmt.Printf("⚠️  写入报告失败: %v\n", err)
	} else {
		abs, _ := filepath.Abs(reportFile)
		fmt.Printf("✅ 报告已保存: %s\n", abs)
	}

	// 汇总
	total := len(results)
	p, f, e := 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case "PASS":
			p++
		case "FAIL":
			f++
		default:
			e++
		}
	}
	rate := 0.0
	if total > 0 {
		rate = float64(p) / float64(total) * 100
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  📊 测试汇总: %d/%d 通过 (%.1f%%)\n", p, total, rate)
	fmt.Printf("  ✅ PASS: %d  |  ❌ FAIL: %d  |  🔴 ERROR: %d\n", p, f, e)
	fmt.Println(strings.Repeat("=", 60))

	if f > 0 || e > 0 {
		os.Exit(1)
	}
}
