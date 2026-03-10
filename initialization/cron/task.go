package cron

import (
	"os"
	"planA/controlState/serviceAlive"
	"planA/controller"
	"planA/initialization/config"
	"planA/modules/logs"
	"planA/modules/pdd"
	"planA/service"
	"planA/tool"
	"time"
)

// DeleteOldRecordsSQLite 删除sqLite中大于7天的记录
func DeleteOldRecordsSQLite() {
	err := service.DeleteOldRecordsSQLite()
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "删除SQLite中7天前的记录失败："+err.Error())
		return
	}
}

// DeleteOldExportSQLite 删除sqLite中大于7天的记录
func DeleteOldExportSQLite() {
	err := service.DeleteOldExportSQLite()
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "删除SQLite中7天前的记录失败："+err.Error())
		return
	}
}

// CheckMysqlAlive mysql心跳
func CheckMysqlAlive() {
	//计算心跳时间
	start := time.Now()
	service.GetTaskUserById(1)
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

// CheckSqliteAlive sqlite心跳
func CheckSqliteAlive() {
	//计算心跳时间
	start := time.Now()
	service.GetTaskRecordById()
	elapsed := time.Since(start)
	elapsedMs := int(elapsed.Milliseconds()) //将time.Duration类型转换为int类型的毫秒
	//设置状态
	serviceAlive.SetServiceAlive("sqlite", elapsedMs)
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

	pddDll, initPddSOErr := pdd.InitPddSO()
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

// DeleteOldExportFile 删除7天前的导出文件
func DeleteOldExportFile() {
	lite, err := service.GetOldExportSQLite()
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取SQLite中7天前的记录失败："+err.Error())
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
