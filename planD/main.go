package main

import (
	"fmt"
	"planA/planD/initialization"
	"planA/planD/logic"
	"planA/planD/validation"
	"time"
)

func main() {

	//校验参数
	taskId, validationErr := validation.Validation()
	if validationErr != nil {
		fmt.Println(validationErr)
		return
	}

	// 初始化
	err := initialization.Init(taskId)
	if err != nil {
		fmt.Println("初始化失败:", err)
		return
	}

	//执行
	err = logic.Logic()
	if err != nil {
		fmt.Println("执行失败:", err)
		return
	}

	// 暂停1分钟，并循环倒计时
	fmt.Println("\n✅ 任务执行完成！")
	fmt.Println("⏸️  暂停1分钟后自动退出...")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := 60; i > 0; i-- {
		minutes := i / 60
		seconds := i % 60
		fmt.Printf("\r⏰ 剩余时间: %02d:%02d", minutes, seconds)
		<-ticker.C
	}
	fmt.Println("\n✨ 程序自动退出")
}
