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
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, configErr.Error())
			return
		}

		// 使用令牌桶进行速率控制（每秒20个）
		if err := golabl.Speed.Wait(golabl.Ctx); err != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("令牌桶等待失败-原因来自于:%v", err))
			continue
		}

		//TODO 重新获取任务头尾
		if taskErr := task.GetTaskHeaderAndFooterSetToG(); taskErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, taskErr.Error())
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
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取任务任务真实的 wait数量失败-原因来自:%v", getTaskBodyWaitCountErr))
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
		if taskPoolErr := golabl.Pool.Pool.Submit(taskExecute); taskPoolErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("协程池意外-原因来自:%d", taskPoolErr))
			golabl.Pool.Wg.Done() // 确保计数正确
		}

		// 判断 任务数是否超过1000 并且 判断是否执行到了1000的倍数
		if golabl.Task.Header.TaskCountTrue > 1000 && golabl.Logic.TaskIndex%1000 == 0 {
			// 更新任务头部信息
			updateTaskHeaderErr := updateTaskHeader()
			if updateTaskHeaderErr != nil {
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("更新任务头信息失败-原因来自:%v", updateTaskHeaderErr))
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
	if updateTaskHeaderErr := updateTaskHeader(); updateTaskHeaderErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("更新任务头信息失败-原因来自:%v", updateTaskHeaderErr))
	}

	// 通知 A程序任务完成
	httpTaskStatusOverErr := tool.NotifyA()
	if httpTaskStatusOverErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, httpTaskStatusOverErr.Error())
	}
}

// 任务执行
func taskExecute() {
	// 任务完成
	defer golabl.Pool.Wg.Done()

	//初始化 变量
	status := golabl.BodyStatusSuccess //默认的书籍执行状态·
	errorStr := "执行成功"                 //默认的书籍执行描述
	// 获取任务信息
	taskMsg, taskMsgErr := service.GetTaskToPopFromBodyWait()
	if errors.Is(taskMsgErr, redis.Nil) {
		//redis 读nil空+1
		fmt.Printf("第 %v 次读出 Redis Nil", atomic.LoadInt64(&golabl.Logic.RedisNilCon))
		atomic.AddInt64(&golabl.Logic.RedisNilCon, 1)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取任务信息失败-原因来自:%v", taskMsgErr))
		return
	} else if taskMsgErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取任务信息失败-原因来自:%v", taskMsgErr))
		return
	}

	// 任务调度
	bodyOverJson, err := dispatcher.Go(taskMsg)
	if err != nil {
		//任务调度失败
		status = golabl.BodyStatusError
		errorStr = fmt.Sprintf("任务调度失败-原因来自:%v", err)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("任务调度失败-原因来自:%v", err))
	} else {
		//任务调度成功
		var bodyOver planAType.TaskBody
		unmarshalErr := json.Unmarshal([]byte(bodyOverJson), &bodyOver)
		if unmarshalErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("bodyOver json.Unmarshal错误-原因:%v", unmarshalErr))
		}
		//更新 taskMsg
		taskMsg = bodyOver
	}

	// 更新任务信息
	taskMsg.Detail.Status = status
	taskMsg.Detail.Error = errorStr

	//isbn 不为空的添加到body中，比如拉取店铺信息isbn可以返回空的
	if taskMsg.BookInfo.Isbn != "" {
		// 添加任务到bodyOver、bodyData、bodyBackup
		if addTaskToBodyOverErr := service.AddTaskToBodyOver(taskMsg); addTaskToBodyOverErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("任务失败 添加到BodyOver失败-原因:%v", addTaskToBodyOverErr))
		}
	}

	// 更新 footer信息
	if updateTaskFooterErr := service.UpdateTaskFooter(status); updateTaskFooterErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("任务失败 添加到BodyOver失败-原因:%v", updateTaskFooterErr))
	}

	// 如果错误是 店铺商品发布达到上限则暂停程序
	if strings.Contains(errorStr, "店铺内发布商品总数已达到上限") {
		golabl.Task.Header.LastIndex = golabl.LastIndexGoodsMaxRestriction
		//暂停 B程序运行
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "任务失败 添加到BodyOver失败-原因:店铺内发布商品总数已达到上限")
		pauseTaskErr := tool.PauseTask()
		if pauseTaskErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "任务失败 添加到BodyOver失败-原因:店铺内发布商品总数已达到上限")
		}
	}

	fmt.Println(errorStr)

}

// 更新头部信息
// @return error 错误信息
func updateTaskHeader() error {
	//通过 footer 来更新 header 的计数
	golabl.Task.Header.TaskCountWait = golabl.Task.Footer.TaskCountWait.Load()
	golabl.Task.Header.TaskCountOver = golabl.Task.Footer.TaskCountOver.Load()
	golabl.Task.Header.TaskCountSuccess = golabl.Task.Footer.TaskCountSuccess.Load()
	golabl.Task.Header.TaskCountError = golabl.Task.Footer.TaskCountError.Load()
	golabl.Task.Header.LastIndex = golabl.Logic.LastIndex
	return service.UpdateTaskHeaderCount()
}
