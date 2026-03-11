package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"planA/modules/logs"
	"planA/planB/config"
	_myRedis "planA/planB/db/redis"
	"planA/planB/dispatcher"
	"planA/planB/dispatcher/pinduoduo"
	"planA/planB/dispatcher/xianYu"
	"planA/planB/interfaces"
	"planA/planB/pool"
	"planA/planB/tool"
	_type "planA/planB/type"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"planA/planB/golabl"

	"golang.org/x/time/rate"

	"sync"
)

var (
	task                _type.Task
	wg                  sync.WaitGroup             // 全局等待组
	globalLimiter       *rate.Limiter              // 全局令牌桶限速器
	lastSecondCount     int64                      // 上一秒处理数量
	lastSecondTimestamp time.Time                  // 上一秒时间戳
	perSecondRateChan   = make(chan float64, 1000) // 每秒速率通道
	// 新增：全局任务监控管理器
	taskMonitors = sync.Map{} // map[string]*TaskMonitor
)

// 新增：任务监控结构体
type TaskMonitor struct {
	redisKey            string
	totalProcessed      int64
	startTime           time.Time
	lastSecondCount     int64
	lastSecondTimestamp time.Time
	lastMinuteCount     int64
	lastFiveMinuteCount int64
	lastFiveMinuteTime  time.Time
	mutex               sync.RWMutex
}

// CreatePlatform 创建平台实例
// @param types 平台类型
// @return platforms.Platform 平台实例
// @return error 错误信息
func CreatePlatform(types string) (interfaces.GoodsTask, error) {
	switch types {
	//case "kongfizi":
	//	return kongFuZi.NewKongfuzi(), nil
	case "pinduoduo":
		return pinDuoDuo.NewPinDuoDuo(), nil
	case "xianyu":
		return xianyu.NewXianYu(), nil
	default:
		return nil, errors.New("错误！")
	}
}
func main() {
	// 截取运行参数中的 任务 ID
	taskKey := os.Args[1]
	if taskKey == "111" {
		//循环1000次
		for i := 0; i < 1000; i++ {
			//每秒打印i
			fmt.Printf("i:%v\n", i)
			time.Sleep(time.Second)
		}
	}
	setConsoleTitle(taskKey)

	// ====================== 初始化 ======================

	// 初始化 CTX
	ctx := context.Background()

	// 初始化令牌桶限速器 (每秒20个请求)
	globalLimiter = rate.NewLimiter(rate.Limit(18), 1)
	log.Printf("令牌桶初始化完成，速率限制: 18次/秒")

	// 初始化配置 并 判断
	mainConfig, configErr := config.Init(ctx, "")
	if configErr != nil {
		errMsg := fmt.Sprintf("初始化配置失败-原因来自于:%v", configErr)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	golabl.MainConfig = &mainConfig
	// 初始化 mysql 并 判断
	//mysqlClient, mysqlErr := mysql.Init(ctx, mainConfig.MysqlConfig)
	//if mysqlErr != nil {
	//	fmt.Printf("初始化 mysql 失败-原因来自于:%v\n", mysqlErr)
	//}
	//defer func() {
	//	if closeErr := mysqlClient.Close(); closeErr != nil {
	//		fmt.Printf("关闭数据库连接失败:%v\n", closeErr)
	//	}
	//}()
	//

	// 初始化 redis（任务池） 并 判断
	redisClientA, redisErr := _myRedis.Init(ctx, mainConfig.RedisConfig[0])
	if redisErr != nil {
		errMsg := fmt.Sprintf("初始化 redis 失败-原因来自于:%v", redisErr)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	// 装入全局变量
	golabl.RedisClientA = redisClientA
	defer func() {
		if closeErr := golabl.RedisClientA.Close(); closeErr != nil {
			errMsg := fmt.Sprintf("关闭 redis 连接失败:%v", closeErr)
			fmt.Printf(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		}
	}()

	// 初始化 redis（出版社信息列表） 并 判断
	redisClientB, redisErr := _myRedis.Init(ctx, mainConfig.RedisConfig[3])
	if redisErr != nil {
		errMsg := fmt.Sprintf("初始化 redis 失败-原因来自于:%v", redisErr)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	// 装入全局变量
	golabl.RedisClientB = redisClientB
	defer func() {
		if closeErr := golabl.RedisClientB.Close(); closeErr != nil {
			errMsg := fmt.Sprintf("关闭 redis 连接失败:%v", closeErr)
			fmt.Printf(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)

		}
	}()

	// 初始化 redis（地区列表） 并 判断
	redisClientC, redisErr := _myRedis.Init(ctx, mainConfig.RedisConfig[4])
	if redisErr != nil {
		errMsg := fmt.Sprintf("初始化 redis 失败-原因来自于:%v", redisErr)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	// 装入全局变量
	golabl.RedisClientC = redisClientC
	defer func() {
		if closeErr := golabl.RedisClientC.Close(); closeErr != nil {
			errMsg := fmt.Sprintf("关闭 redis 连接失败:%v", closeErr)
			fmt.Printf(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)

		}
	}()

	// 初始化 redis（无图isbn） 并 判断
	redisClientD, redisErr := _myRedis.Init(ctx, mainConfig.RedisConfig[5])
	if redisErr != nil {
		errMsg := fmt.Sprintf("初始化 redis 失败-原因来自于:%v", redisErr)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	fmt.Println(mainConfig.RedisConfig[5])
	// 装入全局变量
	golabl.RedisClientD = redisClientD
	//defer func() {
	//	if closeErr := golabl.RedisClientD.Close(); closeErr != nil {
	//		errMsg := fmt.Sprintf("关闭 redis 连接失败:%v", closeErr)
	//		fmt.Printf(errMsg)
	//		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	//
	//	}
	//}()

	// ====================== 初始化 ======================

	// 获取任务头 并 解析 且 判断
	headerErr := _myRedis.GetTaskHeader(golabl.RedisClientA, taskKey, &task.Header)
	if headerErr != nil {
		errMsg := fmt.Sprintf("获取 任务头 失败-原因来自于:%v", headerErr)
		fmt.Println(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}

	// 获取任务尾 并 解析 且 判断
	footerErr := _myRedis.GetTaskFooter(golabl.RedisClientA, taskKey, &task.Footer)
	if footerErr != nil {
		errMsg := fmt.Sprintf("获取 任务尾 失败-原因来自于:%v", footerErr)
		fmt.Println(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}

	// 初始化 协程池
	myPool, poolErr := pool.Init(mainConfig.PoolConfig)
	if poolErr != nil {
		errMsg := fmt.Sprintf("初始化 协程池 失败-原因来自于:%v", poolErr)
		fmt.Println(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	defer func() {
		myPool.Release()
	}()
	/******************************qps******************************/
	//startPerSecondExecutionMonitor(taskKey)
	/****************************qps********************************/

	// 任务循环
	// 统计执行时间 停止
	startBeginz := time.Now()
	taskIndex := 0
	for task.Footer.TaskCountWait.Load() > 0 {
		taskIndex++
		// 使用令牌桶进行速率控制（每秒20个）
		err := globalLimiter.Wait(ctx)
		if err != nil {
			errMsg := fmt.Sprintf("令牌桶等待失败-原因来自于:%v", err)
			log.Printf(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			continue
		}
		// 每次迭代都创建新的变量

		//fmt.Printf("taskIndex:%v %v\n", taskIndex,)
		// 获取任务尾 并 解析 且 判断
		forFooterErr := _myRedis.GetTaskFooter(redisClientA, taskKey, &task.Footer)
		if forFooterErr != nil {
			errMsg := fmt.Sprintf("获取 任务尾 失败-原因来自于:%v", forFooterErr)
			fmt.Println(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		}
		// 判断是否执行到了50的倍数
		if taskIndex%10 == 0 {
			// 获取任务头 并 解析 且 判断
			forHeaderErr := _myRedis.GetTaskHeader(redisClientA, taskKey, &task.Header)
			if forHeaderErr != nil {
				errMsg := fmt.Sprintf("获取 任务头 失败-原因来自于:%v", forHeaderErr)
				fmt.Println(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			}
			// 判断 是否已停止
			if task.Header.Status == _type.TaskStatusStopped {
				fmt.Println("任务已停止~~~~~~~")
				return
			}
		}

		// 创建等待
		wg.Add(1)

		// 协程池 接收
		taskPoolErr := myPool.Submit(func() {
			// 任务完成
			defer wg.Done()

			// 统计执行时间 停止
			//startBegin := time.Now()
			// 初始化 status
			status := int64(1)
			// 初始化 errorStr
			errorStr := "执行成功"
			// 获取任务信息
			taskMsg, taskMsgErr := _myRedis.GetTaskToPopFromBodyWait(golabl.RedisClientA, taskKey)
			if taskMsgErr != nil {
				errMsg := fmt.Sprintf("获取任务信息失败-原因来自:%v", taskMsgErr)
				fmt.Printf(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				return
			}
			if taskMsg.Detail.Price <= 0 {
				status = 2
				errorStr = "价格不能小于0"
			}
			//获取出版社信息并解析
			publishing, getPublishingErr := _myRedis.GetPublishingVid(golabl.RedisClientB, taskMsg.BookInfo.Publishing)
			if getPublishingErr != nil {
				errMsg := fmt.Sprintf("获取出版社信息失败-原因来自:%v", getPublishingErr)
				fmt.Printf(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				return
			}
			taskMsg.Publishing = publishing
			// 统计执行时间 停止
			//timeConsume := time.Since(startBegin).Microseconds()
			//fmt.Printf("任务信息为0:%v;耗时: %d ms\n", taskMsg, timeConsume/1000)

			// 验证违规词
			replaceMark := mainConfig.Server.ReplaceMark
			bannerWordDataReq := map[string]string{
				"isbn":        fmt.Sprintf("%v", taskMsg.BookInfo.Isbn),
				"bookName":    fmt.Sprintf("%v", taskMsg.BookInfo.BookName),
				"author":      fmt.Sprintf("%v", taskMsg.BookInfo.Author),
				"publisher":   fmt.Sprintf("%v", taskMsg.BookInfo.Publishing),
				"shopId":      strconv.FormatInt(task.Header.ShopId, 10),
				"replaceMark": replaceMark,
			}
			var bodyOver _type.TaskBody
			if mainConfig.Server.Filter == 1 {
				substitution, httpBannedWordSubstitutionErr := tool.HttpBannedWordSubstitution(mainConfig.FileUrl.BannedWordSubstitutionUrl, bannerWordDataReq)
				if httpBannedWordSubstitutionErr != nil {
					errMsg := fmt.Sprintf("HttpBannedWordSubstitution 违禁词处理失败-原因来自:%v", httpBannedWordSubstitutionErr)
					fmt.Printf(errMsg)
					logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
					return
				}
				if replaceMark == "0" && len(substitution.Data) > 0 {
					status = 2
					errorStr = "违规词命中 "
					for _, v := range substitution.Data {
						errorStr = errorStr + " " + v.AddTxt + "(" + v.MatchType + ") "
					}
					bodyOver = taskMsg
				}
				if replaceMark == "1" {
					//替换违禁词
					taskMsg.BookInfo.BookName = substitution.BookName
					taskMsg.BookInfo.Author = substitution.Author
					taskMsg.BookInfo.Publishing = substitution.Publisher
					taskMsg.BookInfo.Isbn = substitution.Isbn
				}
			}

			//// 初始化 任务是否错误
			//taskErrBool := true

			// 统计执行时间 停止
			//startBegin1 := time.Now()
			// 进入任务调度
			var bodyOverStr string
			var taskErr error
			var CreatePlatformErr error
			var goodsTask interfaces.GoodsTask
			if status == 1 {
				switch task.Header.ShopType {
				case "1":
					// 初始化 ppdDll
					goodsTask, CreatePlatformErr = CreatePlatform("pinduoduo")
				case "5":
					// 初始化 xianYuDll
					goodsTask, CreatePlatformErr = CreatePlatform("xianyu")
				}
				if CreatePlatformErr != nil {
					configErr := fmt.Errorf("初始化平台失败 %v", err.Error())
					fmt.Println(configErr)
					logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, configErr.Error())
					return
				}
				bodyOverStr, taskErr = dispatcher.Go(goodsTask, "AddGoodsTask", task.Header, taskMsg)
				// 判断错误
				if taskErr != nil {
					status = int64(2)
					errorStr = taskErr.Error()
				}
				unmarshalErr := json.Unmarshal([]byte(bodyOverStr), &bodyOver)
				if unmarshalErr != nil {
					errMsg := fmt.Sprintf("bodyOver json.Unmarshal错误-原因:%v", unmarshalErr)
					fmt.Printf(errMsg)
					logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				}
				// 成功发布后，将商品与店铺的关联关系写入到库中作为商品去重数据

			}
			// 更新任务信息
			bodyOver.Detail.Status = status
			bodyOver.Detail.Error = errorStr
			// 添加任务到 BodyOver
			_, addTaskToBodyOverErr := _myRedis.AddTaskToBodyOver(redisClientA, taskKey, bodyOver)
			if addTaskToBodyOverErr != nil {
				errMsg := fmt.Sprintf("任务失败 添加到BodyOver失败-原因:%v", addTaskToBodyOverErr)
				fmt.Printf(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			}

			// 更新 footer 中的统计信息
			_, updateTaskFooterErr := _myRedis.UpdateTaskFooter(redisClientA, taskKey, &task.Footer, status)
			if updateTaskFooterErr != nil {
				errMsg := fmt.Sprintf("任务失败 添加到BodyOver失败-原因:%v", updateTaskFooterErr)
				fmt.Printf(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			}

			// 统计执行时间 停止
			//timeConsume1 := time.Since(startBegin1).Microseconds()
			//fmt.Printf("任务信息为1:%v;耗时: %d ms\n", taskMsg, timeConsume1/1000)

			if taskErr != nil {
				fmt.Println(taskErr.Error())
			}

			// 关闭当前协程
			return
		})
		if taskPoolErr != nil {
			errMsg := fmt.Sprintf("协程池意外-原因来自:%d", taskPoolErr)
			fmt.Printf(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			wg.Done() // 确保计数正确
		}

		// 判断 任务数是否超过1000
		if task.Header.TaskCountTrue > 1000 {
			// 判断是否执行到了1000的倍数
			if taskIndex%1000 == 0 {
				// 更新任务头部信息
				task.Header.TaskCountWait = task.Footer.TaskCountWait.Load()
				task.Header.TaskCountOver = task.Footer.TaskCountOver.Load()
				task.Header.TaskCountSuccess = task.Footer.TaskCountSuccess.Load()
				task.Header.TaskCountError = task.Footer.TaskCountError.Load()
				_, updateTaskHeaderCountErr := _myRedis.UpdateTaskHeaderCount(redisClientA, taskKey, task.Header)
				if updateTaskHeaderCountErr != nil {
					errMsg := fmt.Sprintf("更新任务尾信息失败-原因来自:%v", updateTaskHeaderCountErr)
					fmt.Printf(errMsg)
					logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				}
			}
		}
	}
	// 等待所有任务完成
	wg.Wait()
	_ = time.Since(startBeginz).Microseconds()
	//fmt.Printf("任务信息 总耗时:%v;耗时: %d ms\n", "", timeConsumez/1000)

	// 更新任务头部信息
	task.Header.TaskCountWait = task.Footer.TaskCountWait.Load()
	task.Header.TaskCountOver = task.Footer.TaskCountOver.Load()
	task.Header.TaskCountSuccess = task.Footer.TaskCountSuccess.Load()
	task.Header.TaskCountError = task.Footer.TaskCountError.Load()
	_, updateTaskHeaderCountErr := _myRedis.UpdateTaskHeaderCount(redisClientA, taskKey, task.Header)
	if updateTaskHeaderCountErr != nil {
		errMsg := fmt.Sprintf("更新任务尾信息失败-原因来自:%v", updateTaskHeaderCountErr)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}

	//TODO
	// 通知 Plan A 任务已完成
	httpTaskStatusOverUrl := mainConfig.HttpUrl.TaskUrl + "/task/over/" + taskKey
	httpCode, httpTaskStatusOverBody, httpTaskStatusOverErr := tool.HttpGetRequest(httpTaskStatusOverUrl)
	if httpTaskStatusOverErr != nil {
		errMsg := fmt.Sprintf("通知A程序任务完成失败-原因来自:%v", httpTaskStatusOverErr)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	// 对通知结果状态进行判断
	var httpTaskStatusOverRes _type.Response
	unmarshalErr := json.Unmarshal([]byte(httpTaskStatusOverBody), &httpTaskStatusOverRes)
	if unmarshalErr != nil {
		errMsg := fmt.Sprintf("通知A程序任务完成失败-原因来自 json.Unmarshal错误: %w %v", unmarshalErr, httpTaskStatusOverBody)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	if httpTaskStatusOverRes.Code != "200" {
		errMsg := fmt.Sprintf("通知A程序任务完成失败-原因来自: url=%v httpCode=%v A程序返回信息=%v\n", httpTaskStatusOverUrl, httpCode, httpTaskStatusOverBody)
		fmt.Printf(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}

}

// 任务检查
func checkTask(mysqlClient *sql.DB, taskMsg _type.TaskBody) error {
	//TODO
	//  // 在 header 中 查看是否需要验证违规
	//  if task.header.IsViolation {
	// 	 // 获取违规信息
	// 	 violation,violationErr := mysql.GetViolation(mysqlClient,taskMsg);if violationErr != nil {
	// 		 return fmt.Errorf("获取违规信息失败")
	// 	 }
	// 	 // 验证违规
	// 	 if mysql.IsViolation(violation) {
	// 		 return fmt.Errorf("任务违规")
	// 	 }
	//  }
	//  // 在 header 中 查看是否需要验证重复
	//  if task.header.IsRepeat {
	// 	 // 获取重复信息
	// 	 repeat,repeatErr := mysql.GetIsRepeat(mysqlClient,taskMsg);if repeatErr != nil {
	// 		 return fmt.Errorf("获取重复信息失败")
	// 	 }
	// 	 // 验证重复
	// 	 if repeat {
	// 		 return fmt.Errorf("商品已存在")
	// 	 }
	return nil
}

func newGetDll(shopType int64) (any, error) {
	//TODO
	// 获取 dll
	// switch shopType {
	// case 0: // 拼多多
	// 	return pinduoduo.Dll,nil
	// case 1: // 孔夫子
	// 	return kongfuzi.Dll,nil
	// case 2: // 咸鱼
	// 	return xianyu.Dll,nil
	// }
	return nil, nil
}

func newGetPlatform(shopType int64) (any, error) {
	//TODO
	// 获取 dll
	// switch shopType {
	// case 0: // 拼多多
	// 	return 	"planA/planB/dispatcher/pinduoduo" ,nil
	// case 1: // 孔夫子
	// 	return kongfuzi.Dll,nil
	// case 2: // 咸鱼
	// 	return xianyu.Dll,nil
	// }
	return nil, nil
}

// 根据店铺类型和任务类型获取令牌桶
func tokenBucketIsAllow(shopType int64, taskType int64) bool {
	//TODO

	return true
}

// 启动每秒速率监控
func startPerSecondRateMonitor() {
	lastSecondTimestamp = time.Now()

	// 每秒计算一次速率
	secondTicker := time.NewTicker(1 * time.Second)
	defer secondTicker.Stop()

	go func() {
		for {
			select {
			case <-secondTicker.C:
				currentTime := time.Now()
				lastSecond := atomic.LoadInt64(&lastSecondCount)

				// 计算每秒速率
				elapsedSeconds := currentTime.Sub(lastSecondTimestamp).Seconds()
				if elapsedSeconds > 0 {
					ratePerSecond := float64(lastSecond) / elapsedSeconds

					// 将速率发送到通道（可用于实时显示或控制）
					select {
					case perSecondRateChan <- ratePerSecond:
					default:
						// 通道满时丢弃旧数据
					}

					// 打印每秒速率（可选）
					log.Printf("当前实时速率: %.2f 条/秒", ratePerSecond)

					// 重置上一秒计数
					atomic.StoreInt64(&lastSecondCount, 0)
					lastSecondTimestamp = currentTime
				}
			}
		}
	}()
}

// 启动每秒执行数量监控 - 支持按redisKey分别统计
func startPerSecondExecutionMonitor(redisKey string) {
	monitor := getOrCreateTaskMonitor(redisKey)
	monitor.startTime = time.Now()
	monitor.lastSecondTimestamp = time.Now()

	lastSecondTotal := int64(0)

	// 每秒打印一次执行数量
	secondTicker := time.NewTicker(1 * time.Second)

	go func() {
		defer secondTicker.Stop()

		for {
			select {
			case <-secondTicker.C:
				// 获取当前任务的监控数据
				monitorNow := getOrCreateTaskMonitor(redisKey)

				currentTotal := atomic.LoadInt64(&monitorNow.totalProcessed)
				processedThisSecond := currentTotal - lastSecondTotal
				lastSecondTotal = currentTotal

				// 计算当前每秒速率
				currentRate := getCurrentRateForTask(redisKey)

				// 计算平均速率
				elapsedSeconds := time.Since(monitorNow.startTime).Seconds()
				avgRatePerSecond := float64(0)
				if elapsedSeconds > 0 {
					avgRatePerSecond = float64(currentTotal) / elapsedSeconds
				}

				// 使用 log.Printf 并添加换行符
				fmt.Printf("====== 每秒执行统计 (任务: %s) ======\n", redisKey)
				fmt.Printf("本秒执行: %d 条\n", processedThisSecond)
				fmt.Printf("当前速率: %.2f 条/秒\n", currentRate)
				fmt.Printf("累计执行: %d 条\n", currentTotal)
				fmt.Printf("平均速率: %.2f 条/秒\n", avgRatePerSecond)
				fmt.Printf("运行时间: %.2f 秒\n", elapsedSeconds)
				fmt.Printf("========================\n")
			}
		}
	}()
}

// 新增：获取或创建任务监控器
func getOrCreateTaskMonitor(redisKey string) *TaskMonitor {
	if monitor, ok := taskMonitors.Load(redisKey); ok {
		return monitor.(*TaskMonitor)
	}

	monitor := &TaskMonitor{
		redisKey:            redisKey,
		startTime:           time.Now(),
		lastSecondTimestamp: time.Now(),
		lastFiveMinuteTime:  time.Now(),
	}

	taskMonitors.Store(redisKey, monitor)
	return monitor
}

// 新增：获取指定任务的当前速率
func getCurrentRateForTask(redisKey string) float64 {
	monitor := getOrCreateTaskMonitor(redisKey)
	currentTime := time.Now()
	lastSecond := atomic.LoadInt64(&monitor.lastSecondCount)

	elapsedSeconds := currentTime.Sub(monitor.lastSecondTimestamp).Seconds()
	if elapsedSeconds < 1 {
		return float64(lastSecond) / math.Max(elapsedSeconds, 0.1)
	}
	return float64(lastSecond) / elapsedSeconds
}

func setConsoleTitle(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleTitle := kernel32.NewProc("SetConsoleTitleW")
	// 将字符串转换为UTF-16指针
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	procSetConsoleTitle.Call(uintptr(unsafe.Pointer(titlePtr)))
}
