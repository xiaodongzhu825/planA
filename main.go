package main

import (
	"fmt"
	"planA/initialization"
	"planA/tool/process"
)

func main() {
	// 初始化
	err := initialization.Init()
	if err != nil {
		fmt.Println("初始化失败:", err)
		return
	}

	//启动C程序
	err = process.RunCProgram()
	if err != nil {
		fmt.Println("启动C程序失败:", err)
		return
	}
	//启动服务
	initialization.Server()
}
