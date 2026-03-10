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
		DeleteOldExportFile()    //删除过期的导出文件
		DeleteOldExportSQLite()  //删除task_export过期记录
		DeleteOldRecordsSQLite() //删除task_record过期记录
	})
	if delSqlIteErr != nil {
		logs.LoggingMiddleware("error", "定时任务 每日执行删除sqlite过期记录 失败")
		return
	}
	//心跳检测 10秒
	_, heartbeatErr := c.AddFunc("0/10 * * * * ?", func() {
		CheckBannedWordSubstitutionUrlAlive() // 违禁词词替换心跳
		CheckMysqlAlive()                     // mysql 心跳
		CheckRedisAlive()                     // redis 心跳
		CheckSqliteAlive()                    // sqlite 心跳
		CheckPddAlive()                       // 拼多多心跳
		CheckCreateTaskNoticeUrlAlive()       // 创建任务通知心跳
		return
	})
	if heartbeatErr != nil {
		logs.LoggingMiddleware("error", "定时任务 心跳检测 失败")
		return
	}

	c.Start() // 启动调度器（非阻塞）
}
