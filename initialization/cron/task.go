package cron

import (
	"fmt"
	"os"
	"planA/controlState/serviceAlive"
	"planA/controller"
	"planA/initialization/config"
	"planA/modules/logs"
	"planA/modules/pdd"
	"planA/rep"
	"planA/service"
	"planA/service/mysql"
	"planA/tool"
	"planA/tool/process"
	"time"
)

// DeleteOldExportFile 删除3天前的导出文件
func DeleteOldExportFile() {
	read := rep.CreateDbFactoryRead()
	lite, getTaskExportOldListErr := read.GetTaskExportOldList()
	if getTaskExportOldListErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取SQLite中3天前的记录失败："+getTaskExportOldListErr.Error())
		return
	}
	for _, v := range lite {
		removeErr := os.Remove(v.FileUrl)
		if removeErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "删除文件失败："+removeErr.Error())
			continue
		}
	}
}

// DeleteOldRecords 删除 task_records 表中3天前的记录
func DeleteOldRecords() {
	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	mysqlDeleteTaskRecordsOldDataErr := mysqlWrite.DeleteTaskRecordsOldData()
	if mysqlDeleteTaskRecordsOldDataErr != nil {
		errMsg := fmt.Sprintf("删除task_records表中3天前的记录失败: %v", mysqlDeleteTaskRecordsOldDataErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
	sqLiteDeleteTaskRecordsOldDataErr := sqliteWrite.DeleteTaskRecordsOldData()
	if sqLiteDeleteTaskRecordsOldDataErr != nil {
		errMsg := fmt.Sprintf("删除task_records表中3天前的记录失败: %v", sqLiteDeleteTaskRecordsOldDataErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
}

// DeleteOldExport 删除  task_export  表中3天前的记录
func DeleteOldExport() {
	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	mysqlDeleteTaskExportOldDataErr := mysqlWrite.DeleteTaskExportOldData()
	if mysqlDeleteTaskExportOldDataErr != nil {
		errMsg := fmt.Sprintf("删除task_export表中3天前的记录失败: %v", mysqlDeleteTaskExportOldDataErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
	sqliteDeleteTaskExportOldDataErr := sqliteWrite.DeleteTaskExportOldData()
	if sqliteDeleteTaskExportOldDataErr != nil {
		errMsg := fmt.Sprintf("删除task_export表中3天前的记录失败: %v", sqliteDeleteTaskExportOldDataErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
}

// CheckMysqlAlive mysql心跳
func CheckMysqlAlive() {
	//计算心跳时间
	start := time.Now()
	mysql.GetTaskRecordsByTaskId("1")
	elapsed := time.Since(start)
	elapsedMs := int(elapsed.Milliseconds()) //将time.Duration类型转换为int类型的毫秒
	//设置状态
	serviceAlive.SetServiceAlive("mysql", elapsedMs)
}

// CheckRedisAlive redis心跳
func CheckRedisAlive() {
	//计算心跳时间
	start := time.Now()
	service.GetTaskBookPing()
	elapsed := time.Since(start)
	elapsedMs := int(elapsed.Milliseconds()) //将time.Duration类型转换为int类型的毫秒
	//设置状态
	serviceAlive.SetServiceAlive("redis", elapsedMs)
}

// CheckPddAlive 拼多多心跳
func CheckPddAlive() {
	token := ""
	//获取系统规定拼多多 token
	//urlConfig, _ := config.GetFileUrlConfig()
	//_, token, HttpGetRequestErr := tool.HttpGetRequest(urlConfig.PddTokenUrl)
	//if HttpGetRequestErr != nil {
	//	logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取系统规定拼多多 token失败："+HttpGetRequestErr.Error())
	//	return
	//}

	pddDll, initPddSOErr := pdd.InitPddDll()
	if initPddSOErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "初始化拼多多dll文件失败："+initPddSOErr.Error())
		return
	}
	client, err := config.GetPddClient()
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取拼多多配置失败："+err.Error())
		return
	}

	//计算心跳时间
	start := time.Now()
	_, pddTimeGetErr := pddDll.PddTimeGet(client.ClientId, client.ClientSecret, token)
	if pddTimeGetErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取拼多多系统时间失败："+pddTimeGetErr.Error())
		return
	}
	elapsed := time.Since(start)
	elapsedMs := int(elapsed.Milliseconds()) //将time.Duration类型转换为int类型的毫秒
	//设置状态
	serviceAlive.SetServiceAlive("pdd", elapsedMs)
}

// CheckCreateTaskNoticeUrlAlive 价软件提交数据通知接口心跳
func CheckCreateTaskNoticeUrlAlive() {
	//计算心跳时间
	start := time.Now()
	controller.TaskNoticeRequest("ping")
	elapsed := time.Since(start)
	elapsedMs := int(elapsed.Milliseconds()) //将time.Duration类型转换为int类型的毫秒
	//设置状态
	serviceAlive.SetServiceAlive("通知取出bodyOver接口", elapsedMs)
}

// CheckBannedWordSubstitutionUrlAlive 违禁词接口心跳
func CheckBannedWordSubstitutionUrlAlive() {

	urlConfig, _ := config.GetFileUrlConfig()
	bannerWordDataReq := map[string]string{
		"isbn":        "9787508618388",
		"bookName":    "麦迪逊大道之王:大卫·奥格威转",
		"author":      "[美]肯尼斯·罗曼",
		"publisher":   "中信出版社",
		"shopId":      "2029141110649929729",
		"replaceMark": "1",
	}
	//计算心跳时间
	start := time.Now()
	tool.HttpBannedWordSubstitution(urlConfig.BannedWordSubstitutionUrl, bannerWordDataReq)
	elapsed := time.Since(start)
	elapsedMs := int(elapsed.Milliseconds()) //将time.Duration类型转换为int类型的毫秒
	//设置状态
	serviceAlive.SetServiceAlive("违禁词替换接口", elapsedMs)
}

// DeleteOldRedis 删除redis3天前的数据
func DeleteOldRedis() {
	read := rep.CreateDbFactoryRead()
	list, getTaskRecordsOldListtErr := read.GetTaskRecordsOldList()
	if getTaskRecordsOldListtErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取task_export中3天前的记录失败："+getTaskRecordsOldListtErr.Error())
		return
	}
	for _, v := range list {
		err := service.DelTask(v.TaskId)
		if err != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "删除任务失败："+err.Error())
			continue
		}

	}
}

// B 程序守护
func B() {
	read := rep.CreateDbFactorySqliteRead()
	//查询task_records中24小时内的所有数据
	records, getTaskRecords24HourErr := read.GetTaskRecords24Hour()
	if getTaskRecords24HourErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取所有任务记录失败："+getTaskRecords24HourErr.Error())
		return
	}
	for _, v := range records {
		//获取 header 信息
		header, getTaskHeaderErr := service.GetTaskHeader(v.TaskId)
		if getTaskHeaderErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取header 信息失败："+getTaskHeaderErr.Error())
			continue
		}
		if header.Status != 0 {
			// 启动 B程序
			_, runTaskWorkerErr := process.RunTaskWorker(v.TaskId)
			if runTaskWorkerErr != nil {
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "启动B程序失败："+runTaskWorkerErr.Error())
				continue
			}
			fmt.Println("守护进程成功启动任务B程序的窗口 任务ID：" + v.TaskId)
		}
	}
}
