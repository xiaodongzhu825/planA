package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	"strings"
	"syscall"
	"time"
	"unsafe"

	"planA/planB/golabl"

	"github.com/go-redis/redis/v8"
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
			//每秒打印 i
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
	// 装入全局变量
	golabl.RedisClientD = redisClientD
	defer func() {
		if closeErr := golabl.RedisClientD.Close(); closeErr != nil {
			errMsg := fmt.Sprintf("关闭 redis 连接失败:%v", closeErr)
			fmt.Printf(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)

		}
	}()

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
	poolConfig := mainConfig.PoolConfig
	if task.Header.Pool.Size > 0 {
		poolConfig = task.Header.Pool
	}

	myPool, poolErr := pool.Init(poolConfig)
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
	redisNilCon := 0      //连续读出 redisNil 的次数
	replaceMarkCon := 0   //连续违规词出现的次数
	lastIndex := int64(0) //记录程序集错误
	for task.Footer.TaskCountWait.Load() > 0 {
		taskIndex++
		//重新获取配置文件
		mainConfig, configErr = config.Init(ctx, "")
		if configErr != nil {
			errMsg := fmt.Sprintf("初始化配置失败-原因来自于:%v", configErr)
			fmt.Printf(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			continue
		}
		golabl.MainConfig = &mainConfig
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

		if redisNilCon > 10 {
			lastIndex = 10001
			//暂停 5000毫秒
			fmt.Printf("连续读出 redisNil 的次数 %v 暂停%v毫秒", redisNilCon, golabl.MainConfig.Server.ErrPauseTime)
			time.Sleep(time.Duration(golabl.MainConfig.Server.ErrPauseTime) * time.Millisecond)
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
			if errors.Is(taskMsgErr, redis.Nil) {
				redisNilCon++
				errMsg := fmt.Sprintf("获取任务信息失败-原因来自:%v", taskMsgErr)
				fmt.Println(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				return
			} else if taskMsgErr != nil {
				errMsg := fmt.Sprintf("获取任务信息失败-原因来自:%v", taskMsgErr)
				fmt.Println(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				return
			}
			if taskMsg.Detail.Price <= 0 {
				status = 2
				errorStr = "价格不能小于0"
			}
			////----------------------处理图片水印----开始-----------------------//
			//if len(taskMsg.BookInfo.ImageObject.CarouselUrlArray) > 0 {
			//	if task.Header.ShopMsg.WatermarkPosition == "1" {
			//		//只给第一张图片添加水印
			//		uploadImage, err := PddUploadImage(taskMsg.BookInfo.ImageObject.CarouselUrlArray[0], task.Header.ShopMsg.WatermarkImgUrl)
			//		if err != nil {
			//			errMsg := fmt.Sprintf("上传图片失败-原因来自:%v", err)
			//			fmt.Printf(errMsg)
			//			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			//		} else {
			//			taskMsg.BookInfo.ImageObject.CarouselUrlArray[0] = uploadImage
			//		}
			//	} else {
			//		//给所有图片添加水印
			//		for i := 0; i < len(taskMsg.BookInfo.ImageObject.CarouselUrlArray); i++ {
			//			uploadImage, err := PddUploadImage(taskMsg.BookInfo.ImageObject.CarouselUrlArray[i], task.Header.ShopMsg.WatermarkImgUrl)
			//			if err != nil {
			//				errMsg := fmt.Sprintf("上传图片失败-原因来自:%v", err)
			//				fmt.Printf(errMsg)
			//				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			//				continue
			//			}
			//			taskMsg.BookInfo.ImageObject.CarouselUrlArray[i] = uploadImage
			//		}
			//	}
			//}
			////----------------------处理图片水印----结束-----------------------//
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
					replaceMarkCon++
					status = 2
					errorStr = "违规词命中 "
					for _, v := range substitution.Data {
						errorStr = errorStr + " " + v.AddTxt + "(" + v.MatchType + ") "
					}
					fmt.Println(errorStr, " isbn：", taskMsg.BookInfo.Isbn)
					bodyOver = taskMsg

					if replaceMarkCon > 10 {
						lastIndex = 10003
						//暂停 B程序运行
						fmt.Println("十次 违规词命中 暂停B程序 运行")
						pauseTaskErr := tool.PauseTask(mainConfig.HttpUrl.TaskUrl, task.Header.TaskId)
						if pauseTaskErr != nil {
							fmt.Println("十次 违规词命中 暂停B程序 运行失败")
						}
						replaceMarkCon = 0
					}
				}
				if replaceMark == "1" {
					//替换违禁词
					taskMsg.BookInfo.BookName = substitution.BookName
					taskMsg.BookInfo.Author = substitution.Author
					taskMsg.BookInfo.Publishing = substitution.Publisher
					taskMsg.BookInfo.Isbn = substitution.Isbn
				}
				replaceMarkCon = 0
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

			// 如果错误是 店铺商品发布达到上限则暂停程序
			if strings.Contains(errorStr, "店铺内发布商品总数已达到上限") {
				lastIndex = 11002
				//暂停 B程序运行
				fmt.Println("店铺内发布商品总数已达到上限 暂停B程序 运行")
				pauseTaskErr := tool.PauseTask(mainConfig.HttpUrl.TaskUrl, task.Header.TaskId)
				if pauseTaskErr != nil {
					fmt.Println("暂停B程序 运行失败")
				}
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
				task.Header.LastIndex = lastIndex
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
	task.Header.LastIndex = lastIndex
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
func setConsoleTitle(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleTitle := kernel32.NewProc("SetConsoleTitleW")
	// 将字符串转换为UTF-16指针
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	procSetConsoleTitle.Call(uintptr(unsafe.Pointer(titlePtr)))
}

//
//// PddUploadImage 拼多多上传打水印后的图片
//func PddUploadImage(img, watermark string) (string, error) {
//	//打水印
//	image, initImageDllErr := image.InitImageDll()
//	if initImageDllErr != nil {
//		fmt.Println("111111111111111")
//		return "", initImageDllErr
//	}
//	ex, addWatermarkFromURLExsErr := image.AddWatermarkFromURLExs(img, watermark)
//	if addWatermarkFromURLExsErr != nil {
//		fmt.Println("222222222222222222")
//		return "", addWatermarkFromURLExsErr
//	}
//	var ret _type.Returns
//	unmarshalErr := json.Unmarshal([]byte(ex), &ret)
//	if unmarshalErr != nil {
//		fmt.Println("---------------------------------------------------------------------------")
//		fmt.Println("img：" + img)
//		fmt.Println("watermark：" + watermark)
//		fmt.Println(ex)
//		fmt.Println("---------------------------------------------------------------------------")
//		return "", unmarshalErr
//	}
//	if !ret.Success {
//		fmt.Println("33333333333333")
//		return "", fmt.Errorf("水印打印失败")
//	}
//	//上传到拼多多
//	pdd, initPddDllErr := pdd.InitPddDll()
//	if initPddDllErr != nil {
//		fmt.Println("4444444444444444")
//		return "", initPddDllErr
//	}
//	upload, pddGoodsImageUploadErr := pdd.PddGoodsImageUpload(ret.Data)
//	if pddGoodsImageUploadErr != nil {
//		fmt.Println("***********************************************************************")
//		fmt.Println(upload)
//		fmt.Println("***********************************************************************")
//		return "", pddGoodsImageUploadErr
//	}
//	if strings.Contains(upload, "错误") {
//		fmt.Println("55555555555555555555")
//		return "", fmt.Errorf(upload)
//	}
//	var goodsImageUploadResponse _type.GoodsImageUploadResponse
//	unmarshalErr = json.Unmarshal([]byte(upload), &goodsImageUploadResponse)
//	if unmarshalErr != nil {
//		fmt.Println("666666666666666666666666666666")
//		return "", unmarshalErr
//	}
//	imgUrl := goodsImageUploadResponse.GoodsImageUploadResponse.ImageURL
//	return imgUrl, nil
//}
