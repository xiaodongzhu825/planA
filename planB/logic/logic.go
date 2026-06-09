package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/planB/dispatcher"
	"planA/planB/initialization/config"
	"planA/planB/initialization/golabl"
	"planA/planB/initialization/task"
	"planA/planB/modules/logs"
	"planA/planB/service"
	"planA/planB/tool"
	planAType "planA/type"
	planATypeMysql "planA/type/mysql"
	redisType "planA/type/redis"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

var Goto bool = false

// Logic 执行任务
func Logic() {
	//loop:
	// 开始读取待处理任务 等待任务数必须大于0
	for golabl.Task.Footer.TaskCountWait.Load() > 0 {
		// 任务索引
		atomic.AddInt64(&golabl.Logic.TaskIndex, 1)

		//TODO 在更新config方法出去后应该去除该代码 每次重新获取配置文件
		if configErr := config.GetConfigSetToG(); configErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, configErr.Error())
			return
		}

		// 使用令牌桶进行速率控制（每秒20个）
		if err := golabl.Speed.Wait(golabl.Ctx); err != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("令牌桶等待失败-原因来自于:%v", err))
			continue
		}

		//TODO 重新获取任务头尾
		if taskErr := task.GetTaskHeaderAndFooterSetToG(); taskErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, taskErr.Error())
			continue
		}

		// 如果连续读出 redisNil 的次数大于100
		if atomic.LoadInt64(&golabl.Logic.RedisNilCon) > 100 {
			//Goto = true

			// 等待所有任务完成 暂停 5 秒
			golabl.Pool.Wg.Wait()
			fmt.Println("等待当前所有协程完成后 暂停5秒，如果等待的任务真的是0的话，则通知A完成任务！")
			time.Sleep(5 * time.Second)

			//获取任务真实的 wait数量
			count, getTaskBodyWaitCountErr := service.GetTaskBodyWaitCount()
			if getTaskBodyWaitCountErr != nil {
				tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取任务任务真实的 wait数量失败-原因来自:%v", getTaskBodyWaitCountErr))
				return
			}
			// 如果数量真的是0，则完成任务
			if count == 0 {
				break
			}

			atomic.StoreInt64(&golabl.Logic.RedisNilCon, 0)
		}

		// 创建等待
		golabl.Pool.Wg.Add(1)

		//协程池 提交
		if golabl.Task.Header.TaskType == 7 {
			// 单线程执行
			taskExecute()
			if taskPoolErr := golabl.Pool.Pool.Submit(taskExecute); taskPoolErr != nil {
				golabl.Pool.Wg.Done()
				tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("协程池意外-原因来自:%d", taskPoolErr))
			} else {
				golabl.Pool.Wg.Done()
			}
		} else {
			// 多线程执行
			if taskPoolErr := golabl.Pool.Pool.Submit(func() {
				defer golabl.Pool.Wg.Done()
				taskExecute()
			}); taskPoolErr != nil {
				golabl.Pool.Wg.Done()
			}
		}

		// 判断 任务数是否超过1000 并且 判断是否执行到了1000的倍数
		if golabl.Task.Header.TaskCountTrue > 1000 && golabl.Logic.TaskIndex%1000 == 0 {
			// 更新任务头部信息
			updateTaskHeaderErr := tool.UpdateTaskHeader()
			if updateTaskHeaderErr != nil {
				tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("更新任务头信息失败-原因来自:%v", updateTaskHeaderErr))
			}
		}
	}

	// 等待所有任务完成
	golabl.Pool.Wg.Wait()

	//等待指定时间后重新执行循环
	//if Goto == true {
	//	golabl.Logic.RedisNilCon = 0
	//	golabl.Logic.LastIndex = golabl.LastIndexRedisNil
	//	fmt.Printf("连续读出 redisNil 的次数 %v 暂停%v毫秒", golabl.Logic.RedisNilCon, golabl.Config.Server.ErrPauseTime)
	//	time.Sleep(time.Duration(golabl.Config.Server.ErrPauseTime) * time.Millisecond)
	//	goto loop
	//}

	// 更新任务头部信息
	if updateTaskHeaderErr := tool.UpdateTaskHeader(); updateTaskHeaderErr != nil {
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("更新任务头信息失败-原因来自:%v", updateTaskHeaderErr))
	}

	// 通知 A程序任务完成
	httpTaskStatusOverErr := tool.NotifyA()
	if httpTaskStatusOverErr != nil {
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, httpTaskStatusOverErr.Error())
	}

	// 延迟2分钟
	time.Sleep(2 * time.Minute)
}

// 任务执行
func taskExecute() {
	//初始化 变量
	status := golabl.BodyStatusSuccess //默认的书籍执行状态·
	errorStr := "执行成功"                 //默认的书籍执行描述

	// 获取任务信息
	taskMsg, taskMsgErr := service.GetTaskToPopFromBodyWait()

	if errors.Is(taskMsgErr, redis.Nil) {
		//redis 读nil空+1
		fmt.Printf("第 %v 次读出 Redis Nil \n", atomic.LoadInt64(&golabl.Logic.RedisNilCon))
		atomic.AddInt64(&golabl.Logic.RedisNilCon, 1)
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取任务信息失败-原因来自:%v", taskMsgErr))
		return
	} else if taskMsgErr != nil {
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取任务信息失败-原因来自:%v", taskMsgErr))
		return
	}

	//设置混合任务成功状态
	if golabl.Task.Header.TaskType == 5 || golabl.Task.Header.TaskType == 9 {
		switch taskMsg.Detail.Status {
		case 1:
			errorStr = "设置商品上架 " + errorStr
			//执行任务
			status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)
		case 2:
			errorStr = "设置商品下架 " + errorStr
			//执行任务
			status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)
		case 3:
			//删除商品的任务存储到 mysql中
			//删除商品 {"book_info":{"isbn":"9787543982888"},"detail":{"goods_id":935670364385,"status":3}}
			DelTask(taskMsg)
			errorStr = "删除商品 已转转移至删除中心"
		case 4:
			errorStr = "修改商品库存 " + errorStr
			//执行任务
			status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)
		case 5:
			errorStr = "修改商品价格 " + errorStr
			//执行任务
			status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)
		case 6:
			errorStr = "发布商品 " + errorStr
			//执行任务
			status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)
		case 7:
			errorStr = "删除并重新发布 " + errorStr
			//执行任务
			status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)

		default:
			//执行任务
			status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)
		}
		// 更新任务信息
		taskMsg.Detail.Status = status
		taskMsg.Detail.Error = errorStr
	} else if golabl.Task.Header.TaskType == 7 {
		//执行任务
		status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)
		taskMsg.Detail.Status = status
		if status != 1 {
			taskMsg.Detail.Error = errorStr
		}
	} else {
		//执行任务
		status, errorStr, taskMsg = exeTask(taskMsg, status, errorStr)
		// 更新任务信息
		taskMsg.Detail.Status = status
		taskMsg.Detail.Error = errorStr
	}

	//isbn 不为空的添加到body中，比如拉取店铺商品信息isbn可以返回空的
	if taskMsg.BookInfo.Isbn != "" && (golabl.TaskType == "3" || golabl.TaskType == "4") {
		// 添加任务到bodyOver、bodyData、bodyBackup
		if addTaskToBodyOverErr := service.AddTaskToBodyOver(taskMsg, []string{}); addTaskToBodyOverErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("任务失败 添加到BodyOver失败-原因:%v", addTaskToBodyOverErr))
		}
	} else {
		if taskMsg.BookInfo.Isbn == "" && taskMsg.BookInfo.BookName == "" {
			taskMsg.BookInfo.BookName = "暂无书品信息"
		}
		// 添加任务到bodyOver、bodyData、bodyBackup
		if addTaskToBodyOverErr := service.AddTaskToBodyOver(taskMsg, []string{}); addTaskToBodyOverErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("任务失败 添加到BodyOver失败-原因:%v", addTaskToBodyOverErr))
		}
	}

	// 更新 footer信息
	if updateTaskFooterErr := service.UpdateTaskFooter(status, 1); updateTaskFooterErr != nil {
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("任务失败 添加到BodyOver失败-原因:%v", updateTaskFooterErr))
	}

	// 如果错误是 店铺商品发布达到上限则暂停程序
	if strings.Contains(errorStr, "店铺内发布商品总数已达到上限") {
		golabl.Task.Header.LastIndex = golabl.LastIndexGoodsMaxRestriction
		//暂停 B程序运行
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "任务失败 添加到BodyOver失败-原因:店铺内发布商品总数已达到上限")
		pauseTaskErr := tool.PauseTask()
		if pauseTaskErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "任务失败 添加到BodyOver失败-原因:店铺内发布商品总数已达到上限")
		}
	}

	fmt.Println(errorStr)
}

//****************************工具**************************************//

// parseShopData 解析店铺数据
// @param shopData 店铺数据
// @return *_type.ShopInfo 店铺信息
func parseShopData(shopData string) (*planAType.ShopInfo, error) {
	shopData = strings.TrimSpace(shopData)

	// 直接解析为 RedisData数组
	var redisData []redisType.RedisData
	err := json.Unmarshal([]byte(shopData), &redisData)
	if err != nil {
		// 尝试另一种格式：可能是单对象而不是数组
		var singleData redisType.RedisData
		if singleErr := json.Unmarshal([]byte(shopData), &singleData); singleErr == nil {
			redisData = []redisType.RedisData{singleData}
		} else {
			return nil, fmt.Errorf("JSON解析失败: %v, 原始数据: %s", err, shopData[:min(100, len(shopData))])
		}
	}

	shopInfo := &planAType.ShopInfo{}

	// 遍历所有数据，根据source_table分类
	for _, item := range redisData {
		switch item.SourceTable {
		case "t_shop":
			var shop planAType.Shop
			if err := json.Unmarshal(item.Data, &shop); err == nil {
				shopInfo.Shop = &shop
			} else {
				fmt.Printf("解析t_shop失败: %v\n", err)
			}
		case "t_shop_detail":
			var detail planAType.ShopDetail
			if err := json.Unmarshal(item.Data, &detail); err == nil {
				shopInfo.ShopDetail = &detail
			} else {
				fmt.Printf("解析t_shop_detail失败: %v\n", err)
			}
		case "t_shop_context":
			var context planAType.ShopContext
			if err := json.Unmarshal(item.Data, &context); err == nil {
				shopInfo.ShopContext = &context
			} else {
				fmt.Printf("解析t_shop_context失败: %v\n", err)
			}
		case "t_spec":
			var spec planAType.Spec
			if err := json.Unmarshal(item.Data, &spec); err == nil {
				shopInfo.Spec = &spec
			} else {
				fmt.Printf("解析t_spec失败: %v\n", err)
			}
		case "t_price_template":
			var template planAType.PriceTemplate
			if err := json.Unmarshal(item.Data, &template); err == nil {
				shopInfo.PriceTemplate = &template
			} else {
				fmt.Printf("解析t_price_template失败: %v\n", err)
			}
		default:
			fmt.Printf("未知的source_table: %s\n", item.SourceTable)
		}
	}

	return shopInfo, nil
}

// 调度任务
func exeTask(taskMsg planAType.TaskBody, status int64, errorStr string) (int64, string, planAType.TaskBody) {
	// 任务调度
	bodyOverJson, err := dispatcher.Go(taskMsg)
	if err != nil {
		//任务调度失败
		status = golabl.BodyStatusError
		errorStr = fmt.Sprintf("任务调度失败-原因来自:%v", err)
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("任务调度失败-原因来自:%v", err))
	} else {
		//任务调度成功
		var bodyOver planAType.TaskBody
		unmarshalErr := json.Unmarshal([]byte(bodyOverJson), &bodyOver)
		if unmarshalErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("bodyOver json.Unmarshal错误-原因:%v", unmarshalErr))
		}
		//更新 taskMsg
		taskMsg = bodyOver
	}
	return status, errorStr, taskMsg
}

// DelTask 删除任务
func DelTask(taskMsg planAType.TaskBody) {
	//删除商品的任务存储到 mysql中
	//删除商品 {"book_info":{"isbn":"9787543982888"},"detail":{"goods_id":935670364385,"status":3}}
	delTask, isExistDelTask, delTaskErr := service.GetDelTaskByTaskId()
	if !isExistDelTask && delTaskErr == nil {
		taskCount := 0
		taskCountOver := 0
		sta := 0
		//将header 转为json
		headerByte, headerJsonErr := json.Marshal(golabl.Task.Header)
		if headerJsonErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("将header 转为json失败-原因来自:%v", headerJsonErr))
			return
		}
		headerJson := string(headerByte)
		currentTime := time.Now()
		// 查询店铺数据
		shopDataStr, getTaskShopErr := service.GetTaskShop(golabl.Task.Header.ShopId)
		if getTaskShopErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("查询店铺数据失败:%v", headerJsonErr))
			return
		}
		// 解析 json数据
		shopData, parseShopDataErr := parseShopData(shopDataStr)
		if parseShopDataErr != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("解析店铺数据失败:%v", parseShopDataErr))
			return
		}
		userId := shopData.Shop.CreateBy
		taskType := 1
		//不存在 mysql任务则创建
		createDelTask := planATypeMysql.DelTask{
			UserID:        &userId,
			ShopID:        &golabl.Task.Header.ShopId,
			TaskID:        &golabl.Task.Header.TaskId,
			ShopName:      &golabl.Task.Header.ShopName,
			ShopType:      &shopData.Shop.ShopType,
			TaskCount:     &taskCount,
			TaskCountOver: &taskCountOver,
			Status:        &sta,
			TaskType:      &taskType,
			Header:        &headerJson,
			CreateAt:      &currentTime,
		}
		var err error
		delTask, err = service.CreateDelTask(createDelTask)
		if err != nil {
			tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("创建删除任务失败-原因来自:%v", err))
			return
		}
	} else if delTaskErr != nil {
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取删除任务失败-原因来自:%v", delTaskErr))
		return
	}

	//将任务状态修改为执行中
	updateDelTaskStatusToDoingErr := service.UpdateDelTaskStatusToDoing()
	if updateDelTaskStatusToDoingErr != nil {
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("将删除任务状态修改为执行中失败-原因来自:%v", updateDelTaskStatusToDoingErr))
		return
	}
	// 将明细的删除任务转移到 mysql中
	insertDelTaskDetailErr := service.InsertDelTaskDetail(delTask.ID, taskMsg)
	if insertDelTaskDetailErr != nil {
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("将明细的删除任务转移到 mysql中失败-原因来自:%v", insertDelTaskDetailErr))
		return
	}
	// 添加删除任务数量
	addDelTaskDetailCountErr := service.AddDelTaskDetailCount()
	if addDelTaskDetailCountErr != nil {
		tool.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("添加删除任务数量失败-原因来自:%v", addDelTaskDetailCountErr))
		return
	}
}
