package initialization

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"planA/planF/initialization/config"
	"planA/planF/initialization/golabl"
	"planA/planF/initialization/router"
)

func Init() error {
	//初始化上下文
	golabl.Ctx = context.Background()
	// 初始化配置
	configErr := config.Init("")
	if configErr != nil {
		return fmt.Errorf("初始化配置失败: %v", configErr)
	}
	//初始化路由
	router.Init()
	return nil
}

// Server 启动服务
func Server() {
	// 从配置获取端口并启动服务
	port := ":" + golabl.Config.Server.FPort
	fmt.Printf("服务器启动在 http://localhost%s\n", port)
	// 打印所有可用端点（控制台输出）
	printAvailableEndpoints()
	// 启动HTTP服务，如果失败则记录致命错误
	log.Fatal(http.ListenAndServe(port, golabl.Router))
}

// printAvailableEndpoints 打印所有可用的API端点
func printAvailableEndpoints() {
	fmt.Println("\n========== 可用API端点 ==========")

	fmt.Println("\n=====================================")
}
