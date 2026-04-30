package cron

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"planA/controlState/serviceAlive"
	"planA/controller"
	"planA/initialization/config"
	"planA/initialization/golabl"
	"planA/modules/logs"
	"planA/modules/pdd"
	"planA/rep"
	"planA/service"
	"planA/service/mysql"
	"planA/tool"
	toolPdd "planA/tool/pdd"
	"planA/tool/process"
	"regexp"
	"strings"
	"time"
)

// DeleteOldExportFile 删除N天前的导出文件
func DeleteOldExportFile() {
	read := rep.CreateDbFactoryRead()
	lite, getTaskExportOldListErr := read.GetTaskExportOldList()
	if getTaskExportOldListErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取SQLite中N天前的记录失败："+getTaskExportOldListErr.Error())
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

// DeleteOldRecords 删除 task_records 表中N天前的记录
func DeleteOldRecords() {
	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	mysqlDeleteTaskRecordsOldDataErr := mysqlWrite.DeleteTaskRecordsOldData()
	if mysqlDeleteTaskRecordsOldDataErr != nil {
		errMsg := fmt.Sprintf("删除task_records表中N天前的记录失败: %v", mysqlDeleteTaskRecordsOldDataErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
	sqLiteDeleteTaskRecordsOldDataErr := sqliteWrite.DeleteTaskRecordsOldData()
	if sqLiteDeleteTaskRecordsOldDataErr != nil {
		errMsg := fmt.Sprintf("删除task_records表中N天前的记录失败: %v", sqLiteDeleteTaskRecordsOldDataErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
}

// DeleteOldExport 删除  task_export  表中N天前的记录
func DeleteOldExport() {
	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	mysqlDeleteTaskExportOldDataErr := mysqlWrite.DeleteTaskExportOldData()
	if mysqlDeleteTaskExportOldDataErr != nil {
		errMsg := fmt.Sprintf("删除task_export表中N天前的记录失败: %v", mysqlDeleteTaskExportOldDataErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
	sqliteDeleteTaskExportOldDataErr := sqliteWrite.DeleteTaskExportOldData()
	if sqliteDeleteTaskExportOldDataErr != nil {
		errMsg := fmt.Sprintf("删除task_export表中N天前的记录失败: %v", sqliteDeleteTaskExportOldDataErr)
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

// DeleteOldRedis 删除redisN天前的数据
func DeleteOldRedis() {
	read := rep.CreateDbFactoryRead()
	list, getTaskRecordsOldListtErr := read.GetTaskRecordsOldList()
	if getTaskRecordsOldListtErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取task_export中N天前的记录失败："+getTaskRecordsOldListtErr.Error())
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

// DeleteOldLog 删除日志N天以上的日志文件
func DeleteOldLog(dir string) {
	// 配置参数
	pattern := `^[^-]+-[a-z]+-(\d{4}-\d{2}-\d{2})(?:-\d{2})?\.log$` // 匹配两种格式：ERROR-task-2026-03-23-04.log 和 ERROR-task-2026-03-23.log
	retentionDays := 3                                              // 保留天数

	// 计算截止时间
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	// 编译正则表达式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("无效的正则表达式模式: %v", err))
		return
	}

	// 遍历目录
	entries, err := os.ReadDir(dir)
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("无法读取目录: %v", err))
		return
	}

	var cleanedCount int
	var errors []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// 检查文件名是否匹配模式并提取日期
		matches := regex.FindStringSubmatch(filename)
		if len(matches) != 2 {
			continue
		}

		// 解析文件名中的日期
		dateStr := matches[1] // 格式: 2026-03-23
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("解析日期失败 %s: %v", filename, err))
			continue
		}

		// 检查文件日期是否早于截止时间
		if fileDate.Before(cutoffTime) {
			filePath := filepath.Join(dir, filename)
			// 删除文件
			if err := os.Remove(filePath); err != nil {
				errors = append(errors, fmt.Sprintf("删除失败 %s: %v", filename, err))
			} else {
				cleanedCount++
			}
		}
	}

	// 输出清理结果
	if cleanedCount > 0 || len(errors) > 0 {
		logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, fmt.Sprintf("清理完成: 删除了 %d 个文件", cleanedCount))
	}

	if len(errors) > 0 {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("清理过程中遇到 %d 个错误", len(errors)))
		for _, errMsg := range errors {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		}
	}
}

// DeleteOldWatermarkImage 删除N天以上的水印图片
func DeleteOldWatermarkImage() {
	// 目标根目录（你提供的目录）
	rootDir := `img\watermark`

	// 计算需要删除的截止时间：当前时间 - N天
	days := golabl.Config.Server.DataDay
	expireTime := time.Now().AddDate(0, 0, -days)

	// 遍历根目录
	err := filepath.Walk(rootDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("无法访问目录 %s: %v", path, err)
		}

		// 只处理一级子文件夹（不递归）
		if path == rootDir {
			return nil
		}
		if !f.IsDir() {
			return nil
		}

		// 解析文件夹名称为日期（格式：2006-01-02）
		dirName := f.Name()
		dirTime, err := time.Parse("2006-01-02", dirName)
		if err != nil {
			// 不是日期格式的文件夹跳过
			return nil
		}

		// 判断是否超过N天
		if dirTime.Before(expireTime) {
			// 删除文件夹（包括里面所有内容）
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("无法删除目录 %s: %v", path, err)
			}
		}

		// 只处理一级子目录，不递归深入
		return filepath.SkipDir
	})
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("删除N天以上的水印图片失败: %v", err.Error()))
	}
}

// DeleteOldSkuWatermarkImage 删除N天以上的sku水印图片
func DeleteOldSkuWatermarkImage() {
	// 目标根目录（你提供的目录）
	rootDir := `img\skuwatermark`

	// 计算需要删除的截止时间：当前时间 - N天
	days := golabl.Config.Server.DataDay
	expireTime := time.Now().AddDate(0, 0, -days)

	// 遍历根目录
	err := filepath.Walk(rootDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("无法访问目录 %s: %v", path, err)
		}

		// 只处理一级子文件夹（不递归）
		if path == rootDir {
			return nil
		}
		if !f.IsDir() {
			return nil
		}

		// 解析文件夹名称为日期（格式：2006-01-02）
		dirName := f.Name()
		dirTime, err := time.Parse("2006-01-02", dirName)
		if err != nil {
			// 不是日期格式的文件夹跳过
			return nil
		}

		// 判断是否超过N天
		if dirTime.Before(expireTime) {
			// 删除文件夹（包括里面所有内容）
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("无法删除目录 %s: %v", path, err)
			}
		}

		// 只处理一级子目录，不递归深入
		return filepath.SkipDir
	})
	if err != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("删除N天以上的水印图片失败: %v", err.Error()))
	}
}

// BackupBodyBackup 备份 body_backup到硬盘
func BackupBodyBackup() {

	// 定义导出目录
	csvUrl := golabl.Config.FileUrl.BackupUrl
	fmt.Println("路径：" + csvUrl)
	// 获取所有任务数据
	read := rep.CreateDbFactoryRead()
	list, getAllTaskErr := read.GetAllTask()
	if getAllTaskErr != nil {
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, "获取所有任务数据失败："+getAllTaskErr.Error())
		return
	}

	for _, v := range list {

		dateDir := v.CreateAt.Format("2006-01-02") // 按日期分组
		// 构建完整的目录路径
		taskCsvUrl := filepath.Join(csvUrl, dateDir)

		// 获取 backup数据长度
		backupLen, getBodyBackupLenErr := service.GetBodyBackupLen(v.TaskId)
		if getBodyBackupLenErr != nil {
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取任务 %s 的backup长度失败：%v", v.TaskId, getBodyBackupLenErr))
			continue // 跳过当前任务，继续处理下一个
		}
		if backupLen == 0 {
			// 如果 backup中没有数据则跳过
			continue
		}

		csvFileName := fmt.Sprintf("%v.csv", v.TaskId)
		// 检查并创建目录（如果不存在）
		err := os.MkdirAll(taskCsvUrl, 0755)
		if err != nil {
			errMsg := fmt.Sprintf("创建目录失败: %v", err)
			fmt.Println(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			return
		}

		// 拼接完整的文件路径
		fullPath := filepath.Join(taskCsvUrl, csvFileName)

		// 判断文件是否存在，决定是创建新文件还是追加写入
		var file *os.File
		if _, err := os.Stat(fullPath); err == nil {
			// 文件存在，以追加模式打开
			file, err = os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				errMsg := fmt.Sprintf("打开文件失败: %v", err)
				fmt.Println(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				continue
			}
		} else if os.IsNotExist(err) {
			// 文件不存在，创建新文件
			file, err = os.Create(fullPath)
			if err != nil {
				errMsg := fmt.Sprintf("创建文件失败: %v", err)
				fmt.Println(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				continue
			}
		} else {
			// 其他错误（如权限问题）
			errMsg := fmt.Sprintf("检查文件状态失败: %v", err)
			fmt.Println(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			continue
		}
		defer file.Close()

		// 创建 CSV写入器
		writer := csv.NewWriter(file)
		defer writer.Flush()

		// 循环获取并写入数据
		for i := 0; i < int(backupLen); i++ {
			// 获取 backup数据
			one, getBodyBackupOneErr := service.GetBodyBackupOne(v.TaskId)
			if getBodyBackupOneErr != nil {
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("获取任务 %s 的body_backup数据失败：%v", v.TaskId, getBodyBackupOneErr))
				break // 跳出当前循环，继续下一个任务
			}

			// 将数据写入到CSV文件的一行（A列）
			// 假设 one 是字符串类型，如果是结构体需要根据实际字段调整
			record := []string{one} // 如果 one 不是字符串，需要转换，例如 fmt.Sprintf("%v", one)

			if err := writer.Write(record); err != nil {
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, fmt.Sprintf("写入CSV数据失败：%v", err))
				break
			}
		}

		logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, fmt.Sprintf("任务 %s 的数据已成功写入文件：%s", v.TaskId, fullPath))
	}
}

// ZipBackupFile 压缩backup文件
func ZipBackupFile() {
	csvUrl := golabl.Config.FileUrl.BackupUrl
	// 获取昨天的日期
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	// 拼接完整的文件路径
	taskCsvUrl := filepath.Join(csvUrl, yesterday)

	// 检查源目录是否存在
	srcInfo, err := os.Stat(taskCsvUrl)
	if os.IsNotExist(err) {
		log.Printf("目录不存在: %s", taskCsvUrl)
		return
	}

	// 如果不是目录，则直接返回
	if !srcInfo.IsDir() {
		log.Printf("路径不是目录: %s", taskCsvUrl)
		return
	}

	// 创建zip文件
	zipFileName := taskCsvUrl + ".zip"
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		log.Printf("创建zip文件失败: %v", err)
		return
	}
	defer zipFile.Close()

	// 创建zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历目录并添加文件
	err = filepath.Walk(taskCsvUrl, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if info.IsDir() {
			return nil
		}

		// 获取相对路径作为zip内的文件名
		relPath, err := filepath.Rel(taskCsvUrl, path)
		if err != nil {
			return err
		}

		// 创建zip中的文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.Join(yesterday, relPath) // 保留目录结构
		header.Method = zip.Deflate

		// 创建zip中的文件
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 打开并复制源文件
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		log.Printf("压缩失败: %v", err)
		os.Remove(zipFileName) // 删除不完整的zip文件
		return
	}

	log.Printf("压缩成功: %s", zipFileName)

	// 压缩成功后删除原目录
	err = os.RemoveAll(taskCsvUrl)
	if err != nil {
		log.Printf("删除原目录失败: %v", err)
		return
	}
	log.Printf("成功删除原目录: %s", taskCsvUrl)
}

// DeleteZipFile 删除zip文件
func DeleteZipFile() {
	zipDir := golabl.Config.FileUrl.BackupUrl
	day := golabl.Config.Server.DataDay

	// 获取当前时间
	now := time.Now()
	// 计算截止时间（当天0点减去指定天数）
	cutoffTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -day)

	// 读取目录
	entries, err := os.ReadDir(zipDir)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return
	}

	deletedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".zip") {
			continue
		}

		// 从文件名解析日期（格式：2026-04-10.zip）
		dateStr := strings.TrimSuffix(fileName, ".zip")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// 如果文件名不符合日期格式，跳过
			continue
		}

		// 如果文件日期早于或等于截止时间，则删除
		if !fileDate.After(cutoffTime) {
			filePath := filepath.Join(zipDir, fileName)
			err := os.Remove(filePath)
			if err != nil {
				fmt.Printf("删除文件失败 %s: %v\n", filePath, err)
			} else {
				fmt.Printf("已删除文件: %s\n", filePath)
				deletedCount++
			}
		}
	}

	fmt.Printf("共删除 %d 个%d天前的zip文件\n", deletedCount, day)
}

// ExecuteDelTask 查询del_task 表中待执行
func ExecuteDelTask() {
	delTask, err := mysql.GetDelTask()
	if err != nil {
		fmt.Println("查询del_task 表中待执行失败：", err)
	}
	for _, v := range delTask {
		//如果是暂停中的任务,将任务状态修改为执行中
		if *v.Status == 2 {
			// 要求 v.PauseAt 不能等于 nil 并且 v.PauseAt 必须大于当前时间24小时
			if v.PauseAt != nil && !time.Now().After(v.PauseAt.Add(24*time.Hour)) {
				continue
			}
			//修改任务状态
			err := mysql.UpdateDelTaskStatus(v.ID)
			if err != nil {
				fmt.Println("修改任务状态失败：", err)
				continue
			}
		}
		//启动任务
		_, err := process.RunDprogram(*v.TaskID)
		if err != nil {
			fmt.Println("启动任务失败：", err)
			continue
		}
	}
}

// DeleteDelTaskAndDelTaskDetails 清理删除任务与删除任务详情过期的数据
func DeleteDelTaskAndDelTaskDetails() {
	task, getExpiredDelTaskErr := mysql.GetExpiredDelTask()
	if getExpiredDelTaskErr != nil {
		fmt.Println("查询过期的删除任务失败：", getExpiredDelTaskErr)
		return
	}
	for _, v := range task {
		if *v.TaskType == 2 || *v.TaskType == 3 {
			//删除header 与 footer
			delTaskErr := service.DelTask(*v.TaskID)
			if delTaskErr != nil {
				fmt.Println("删除任务失败：", delTaskErr)
				return
			}
		}
		// 删除删除任务明细表
		deleteDelTaskDetailErr := mysql.DeleteDelTaskDetail(*v.TaskID)
		if deleteDelTaskDetailErr != nil {
			fmt.Println("删除删除任务明细表失败：", deleteDelTaskDetailErr)
			return
		}
		//删除任务
		deleteDelTaskByIdErr := mysql.DeleteDelTaskById(v.ID)
		if deleteDelTaskByIdErr != nil {
			fmt.Println("删除任务失败：", deleteDelTaskByIdErr)
			return
		}
	}
}

// VerifyToken 验证token过期

func VerifyToken() {
	list, getPddTokenListErr := service.GetPddTokenList()
	if getPddTokenListErr != nil {
		fmt.Println("获取token列表失败：", getPddTokenListErr)
		return
	}
	pddDll, initPddSOErr := pdd.InitPddDll()
	if initPddSOErr != nil {
		fmt.Println("初始化pdd.so失败：", initPddSOErr)
		return
	}
	for _, v := range list {
		//使用类目预测接口测试Token 是否有效
		buildPddGoodsOuterCatMappingGetErr := toolPdd.BuildPddGoodsOuterCatMappingGet(pddDll, v.Token)
		if buildPddGoodsOuterCatMappingGetErr != nil {
			if buildPddGoodsOuterCatMappingGetErr.Error() == "拼多多Token已过期" {
				fmt.Printf("token 过期的店铺 %v 店铺id %v\n", v.ShopName, v.ID)
				reqData := map[string]string{
					"shopId": v.ID,
				}
				_, submitFormDataErr := tool.SubmitFormData(golabl.Config.FileUrl.UpdateTokenUrl, reqData)
				if submitFormDataErr != nil {
					fmt.Println("提交表单数据失败：", submitFormDataErr)
					return
				}
			}
		}
	}
}
