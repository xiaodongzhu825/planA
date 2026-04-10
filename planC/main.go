package main

import (
	"context"
	"fmt"
	"planA/initialization/config"
	planACron "planA/initialization/cron"
	"planA/initialization/golabl"
	"planA/initialization/mysql"
	"planA/initialization/redis"
	"planA/initialization/sqLite"
	"sync"
)

var (
	backupMutex sync.Mutex
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
	return nil
}

func main() {

	// 初始化
	err := Init()
	if err != nil {
		fmt.Println("初始化失败:", err)
		return
	}

	fmt.Println("定时任务 启动成功")

	fmt.Println("开始备份 body_backup到硬盘")
	planACron.BackupBodyBackup()
	fmt.Println("备份 body_backup到硬盘完成")

	//
	//c := cron.New(cron.WithSeconds()) // 支持秒级别的精度
	//
	//// 备份 body_backup到硬盘 - 每分钟执行一次，使用锁防止并发
	//_, backupBodyBackupErr := c.AddFunc("0 * * * * ?", func() {
	//	// 尝试获取锁，如果锁已被占用则直接返回
	//	if !backupMutex.TryLock() {
	//		fmt.Println("上一次备份任务尚未完成，跳过本次执行")
	//		return
	//	}
	//	defer backupMutex.Unlock()
	//	fmt.Println("开始备份 body_backup到硬盘")
	//	planACron.BackupBodyBackup()
	//	fmt.Println("备份 body_backup到硬盘完成")
	//
	//})
	//if backupBodyBackupErr != nil {
	//	fmt.Println("定时任务 备份 body_backup到硬盘 启动失败")
	//	return
	//}
	//
	//// 每天上午9点压缩昨天csv文件
	//_, zipBackupFileErr := c.AddFunc("0 0 9 * * ?", func() {
	//	fmt.Println("开始压缩昨天 csv文件")
	//	planACron.ZipBackupFile()
	//})
	//if zipBackupFileErr != nil {
	//	fmt.Println("定时任务 zipBackupFile 启动失败")
	//	return
	//}
	//
	//c.Run() // 启动调度器（阻塞运行）
}
