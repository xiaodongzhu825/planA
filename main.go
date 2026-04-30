package main

import (
	"fmt"
	"planA/initialization"
)

func main() {
	// 初始化
	err := initialization.Init()
	if err != nil {
		fmt.Println("初始化失败:", err)
		return
	}
	//cron.VerifyToken()
	//启动服务
	initialization.Server()
}
