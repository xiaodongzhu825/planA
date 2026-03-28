package main

import (
	"fmt"
	"planA/planB/initialization"
	"planA/planB/initialization/golabl"
	"planA/planB/logic"
	"planA/planB/modules/logs"
	"planA/planB/tool"
	"planA/planB/validation"
	"time"
)

func main() {

	//校验参数
	taskId, validationErr := validation.Validation()
	if validationErr != nil {
		fmt.Println(validationErr)
		return
	}

	// 是否测试模式
	if taskId == "111" {
		test()
		return
	}

	// 初始化配置
	err := initialization.Init(taskId)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 拉取商品列表
	if golabl.Task.Header.TaskType == 3 {
		golabl.Platform.GetGoodsTask()
		// 通知 A程序任务完成
		httpTaskStatusOverErr := tool.NotifyA()
		if httpTaskStatusOverErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, httpTaskStatusOverErr.Error())
		}
		//延迟3分钟,并且循环打印每秒倒计时
		totalSeconds := 180 // 3分钟 = 180秒
		for i := totalSeconds; i >= 0; i-- {
			minutes := i / 60
			seconds := i % 60
			fmt.Printf("\r剩余时间: %02d:%02d", minutes, seconds)
			if i > 0 {
				time.Sleep(1 * time.Second)
			}
		}
	} else {
		// 执行任务
		logic.Logic()
	}

}

// 测试模式
func test() {
	//循环1000次
	for i := 0; i < 1000; i++ {
		//每秒打印 i
		fmt.Printf("i:%v\n", i)
		time.Sleep(time.Second)
	}
}
