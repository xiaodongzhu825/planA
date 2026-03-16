package initialization

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"planA/initialization/config"
	"planA/initialization/cron"
	"planA/initialization/golabl"
	"planA/initialization/middle"
	"planA/initialization/mysql"
	"planA/initialization/redis"
	"planA/initialization/router"
	"planA/initialization/sqLite"
	"planA/initialization/validator"
)

func Init() error {
	//初始化上下文
	golabl.Ctx = context.Background()
	// 初始化配置
	configErr := config.Init("")
	if configErr != nil {
		return fmt.Errorf("初始化配置失败: %v", configErr)
	}
	// 初始化 mysql
	mysqlErr := mysql.Init()
	if mysqlErr != nil {
		return fmt.Errorf("初始化mysql失败: %v", mysqlErr)
	}
	// 初始化 redis
	redisErr := redis.Init()
	if redisErr != nil {
		return fmt.Errorf("初始化redis失败: %v", redisErr)
	}
	// 初始化 sqlite
	sqliteErr := sqLite.Init()
	if sqliteErr != nil {
		return fmt.Errorf("初始化sqlite失败: %v", sqliteErr)
	}
	// 初始化验证器
	validator.Init()
	// 初始化定时任务（非阻塞，因此不需要返回错误）
	cron.Init()
	//初始化中间件
	middle.Init()
	//初始化路由
	router.Init()
	return nil

}

// Server 启动服务
func Server() {
	// 从配置获取端口并启动服务
	port := ":" + golabl.Config.Server.Port
	fmt.Printf("服务器启动在 http://localhost%s\n", port)
	// 打印所有可用端点（控制台输出）
	printAvailableEndpoints()
	// 启动HTTP服务，如果失败则记录致命错误
	log.Fatal(http.ListenAndServe(port, golabl.Router))
}

// printAvailableEndpoints 打印所有可用的API端点
func printAvailableEndpoints() {
	fmt.Println("\n========== 可用API端点 ==========")

	fmt.Println("\n【任务管理】")
	fmt.Println("  POST   /task/create                     - 创建新任务")
	fmt.Println("  GET    /task/pause/{id}                  - 暂停任务")
	fmt.Println("  GET    /task/resume/{id}                 - 恢复任务")
	fmt.Println("  GET    /task/stop/{id}                   - 停止任务")
	fmt.Println("  GET    /task/over/{id}                   - 完成任务")
	fmt.Println("  GET    /task/get                         - 获取任务列表（支持查询参数）")
	fmt.Println("  GET    /task/getByUserId                  - 根据用户ID获取任务")
	fmt.Println("  POST   /task/setTaskBody                  - 设置任务内容")
	fmt.Println("  GET    /task/b                            - 运行B程序")

	fmt.Println("\n【任务导出】")
	fmt.Println("  GET    /task/export/exportTaskDetail/{id} - 导出指定任务详情")
	fmt.Println("  GET    /task/export/exportTaskDetail/{userId}/{id} - 导出指定用户的指定任务详情")
	fmt.Println("  GET    /task/export/get                   - 获取所有导出任务列表")
	fmt.Println("  GET    /task/export/get/{userId}          - 获取指定用户的导出任务列表")

	fmt.Println("\n【商品任务】")
	fmt.Println("  POST   /task/goods/add                     - 添加商品任务")
	fmt.Println("  GET    /task/goods/get/{id}                - 获取指定商品任务详情")
	fmt.Println("  PUT    /task/goods/set/{id}                 - 更新指定商品任务")
	fmt.Println("  DELETE /task/goods/del/{id}                - 删除指定商品任务")

	fmt.Println("\n【系统工具】")
	fmt.Println("  GET    /alive/get                           - 获取服务存活状态列表")
	fmt.Println("  GET    /health                              - 健康检查")
	fmt.Println("  GET    /export/                             - 导出文件下载服务")
	fmt.Println("  GET    /                                    - 服务欢迎页")

	fmt.Println("\n=====================================")
}
