package cron

import (
	"planA/modules/logs"

	"github.com/robfig/cron/v3"
)

// Init 定时器初始化
func Init() {
	c := cron.New(cron.WithSeconds()) // 支持秒级别的精度
	// 每日执行删除sqlite过期记录
	_, delSqlIteErr := c.AddFunc("0 0 0 * * ?", func() {
		DeleteOldSkuWatermarkImage() //删除过期的 sku水印图片
		DeleteOldWatermarkImage()    //删除过期的水印图片
		DeleteOldExportFile()        //删除过期的导出文件
		DeleteOldRedis()             //删除 redis中过期数据
		DeleteOldRecords()           //删除task_record过期记录
		DeleteOldExport()            //删除task_export过期记录
		DeleteZipFile()              //删除 zip文件
	})
	if delSqlIteErr != nil {
		logs.LoggingMiddleware("error", "定时任务 每日执行删除sqlite过期记录 失败")
		return
	}
	//心跳检测 10秒
	_, heartbeatErr := c.AddFunc("0/10 * * * * ?", func() {
		CheckBannedWordSubstitutionUrlAlive() // 违禁词替换心跳
		CheckMysqlAlive()                     // mysql 心跳
		CheckRedisAlive()                     // redis 心跳
		CheckPddAlive()                       // 拼多多心跳
		CheckCreateTaskNoticeUrlAlive()       // 创建任务通知心跳
		return
	})
	if heartbeatErr != nil {
		logs.LoggingMiddleware("error", "定时任务 心跳检测 失败")
		return
	}
	// 60秒钟检测一次
	_, bErr := c.AddFunc("0/60 * * * * ?", func() {
		B()
	})
	if bErr != nil {
		logs.LoggingMiddleware("error", "定时任务 B 函数 启动失败")
		return
	}

	// 每日执行删除过期日志文件
	_, delLogErr := c.AddFunc("0 0 0 * * ?", func() {
		DeleteOldLog("logs\\debug")
		DeleteOldLog("logs\\info")
		DeleteOldLog("logs\\warning")
		DeleteOldLog("logs\\error")
		DeleteOldLog("logs\\success")
	})
	if delLogErr != nil {
		logs.LoggingMiddleware("error", "定时任务 删除过期日志文件 启动失败")
		return
	}

	//// 备份 body_backup到硬盘 - 每分钟执行一次，使用锁防止并发（挪到C.exe）
	//_, backupBodyBackupErr := c.AddFunc("0 * * * * ?", func() {
	//	BackupBodyBackup()
	//})
	//if backupBodyBackupErr != nil {
	//	logs.LoggingMiddleware("error", "定时任务 备份 body_backup到硬盘 启动失败")
	//	return
	//}
	//
	//// 每天上午9点压缩昨天csv文件（挪到C.exe）
	//_, zipBackupFileErr := c.AddFunc("0 0 9 * * ?", func() {
	//	ZipBackupFile()
	//})
	//if zipBackupFileErr != nil {
	//	logs.LoggingMiddleware("error", "定时任务 zipBackupFile 启动失败")
	//	return
	//}

	c.Start() // 启动调度器（非阻塞）
}
