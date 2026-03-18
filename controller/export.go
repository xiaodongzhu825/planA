package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"planA/modules/logs"
	"planA/rep"
	"planA/service"
	"planA/tool"
	_type "planA/type"
	"planA/validator"
)

// GetExportTask 导出任务列表
func GetExportTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, GetExportValidatorErr := validator.GetExportValidator(data)
	if GetExportValidatorErr != nil {
		tool.Error(httpMsg, GetExportValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)

	read := rep.CreateDbFactoryRead()
	records, total, getTaskRecordsListErr := read.GetTaskExportList(page, size, "")
	if getTaskRecordsListErr != nil {
		errMsg := getTaskRecordsListErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	var dataTaskAll []map[string]interface{}
	for _, v := range records {
		complete, getExportFileProgressErr := service.GetExportFileProgress(v.TaskId)
		if getExportFileProgressErr != nil {
			errMsg := "获取任务进度失败: " + getExportFileProgressErr.Error()
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			continue
		}
		taskExportdata := map[string]interface{}{
			"task_id":     v.TaskId,
			"shop_name":   v.ShopName,
			"status":      v.Status,
			"total":       v.Total,
			"file_url":    v.FileUrl,
			"complete_at": v.CompleteAt.Time,
			"create_at":   v.CreateAt,
			"complete":    complete,
		}
		dataTaskAll = append(dataTaskAll, taskExportdata)
	}
	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  dataTaskAll,
	}
	tool.Session(httpMsg, dataRet)
}

// GetExportTaskByUserId 导出任务列表-用户
func GetExportTaskByUserId(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, GetExportByUserIdValidatorErr := validator.GetExportByUserIdValidator(data)
	if GetExportByUserIdValidatorErr != nil {
		tool.Error(httpMsg, GetExportByUserIdValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)

	read := rep.CreateDbFactoryRead()
	records, total, getTaskRecordsListErr := read.GetTaskExportList(page, size, dataVal.UserID)
	if getTaskRecordsListErr != nil {
		errMsg := getTaskRecordsListErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	dataTaskAll := []map[string]interface{}{}
	for _, v := range records {
		complete, getExportFileProgressErr := service.GetExportFileProgress(v.TaskId)
		if getExportFileProgressErr != nil {
			errMsg := "获取任务进度失败: " + getExportFileProgressErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		taskExportdata := map[string]interface{}{
			"task_id":     v.TaskId,
			"shop_name":   v.ShopName,
			"status":      v.Status,
			"total":       v.Total,
			"file_url":    v.FileUrl,
			"complete_at": v.CompleteAt.Time,
			"create_at":   v.CreateAt,
			"complete":    complete,
		}
		dataTaskAll = append(dataTaskAll, taskExportdata)
	}
	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  dataTaskAll,
	}
	tool.Session(httpMsg, dataRet)
}

// ExportTaskDetail 根据任务 id导出任务详情
func ExportTaskDetail(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, GetExportDetailValidatorErr := validator.GetExportDetailValidator(data)
	if GetExportDetailValidatorErr != nil {
		tool.Error(httpMsg, GetExportDetailValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	read := rep.CreateDbFactoryRead()

	//查询是任务信息
	taskRecord, getTaskRecordsByTaskIDErr := read.GetTaskRecordsByTaskId(dataVal.TaskID)
	if getTaskRecordsByTaskIDErr != nil {
		errMsg := fmt.Sprintf("获取任务信息失败 %v", getTaskRecordsByTaskIDErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	if taskRecord.IsExport == 1 {
		errMsg := "任务已导出过"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//获取任务详情总数
	total, GetBodyOverCount := service.GetBodyOverCount(dataVal.TaskID)
	if GetBodyOverCount != nil {
		errMsg := fmt.Sprintf("获取任务详情总数失败 %v", GetBodyOverCount)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	//创建一条导出任务
	var status int64
	var fileUrl string
	mysqlCreateTaskExportErr := mysqlWrite.CreateTaskExport(_type.TaskExportDTO{
		UserId:     taskRecord.UserId,
		ShopId:     taskRecord.ShopId,
		TaskId:     taskRecord.TaskId,
		ShopName:   taskRecord.ShopName,
		FileUrl:    fileUrl,
		Status:     status,
		Total:      total,
		CompleteAt: sql.NullTime{},
	})
	if mysqlCreateTaskExportErr != nil {
		errMsg := fmt.Sprintf("写入任务信息失败 %v", mysqlCreateTaskExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	sqLiteCreateTaskExportErr := sqliteWrite.CreateTaskExport(_type.TaskExportDTO{
		UserId:     taskRecord.UserId,
		ShopId:     taskRecord.ShopId,
		TaskId:     taskRecord.TaskId,
		ShopName:   taskRecord.ShopName,
		FileUrl:    fileUrl,
		Status:     status,
		Total:      total,
		CompleteAt: sql.NullTime{},
	})
	if sqLiteCreateTaskExportErr != nil {
		errMsg := fmt.Sprintf("写入任务信息失败 %v", sqLiteCreateTaskExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//修改任务导出状态
	mysqlUpdateTaskRecordsErr := mysqlWrite.UpdateTaskRecords(_type.TaskRecordsDTO{
		Id:       taskRecord.Id,
		UserId:   taskRecord.UserId,
		ShopId:   taskRecord.ShopId,
		TaskId:   taskRecord.TaskId,
		ShopName: taskRecord.ShopName,
		IsExport: 1,
		TaskType: taskRecord.TaskType,
	})
	if mysqlUpdateTaskRecordsErr != nil {
		errMsg := fmt.Sprintf("修改任务导出状态失败 %v", mysqlUpdateTaskRecordsErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	sqLiteUpdateTaskRecordsErr := sqliteWrite.UpdateTaskRecords(_type.TaskRecordsDTO{
		Id:       taskRecord.Id,
		UserId:   taskRecord.UserId,
		ShopId:   taskRecord.ShopId,
		TaskId:   taskRecord.TaskId,
		ShopName: taskRecord.ShopName,
		IsExport: 1,
		TaskType: taskRecord.TaskType,
	})
	if sqLiteUpdateTaskRecordsErr != nil {
		errMsg := fmt.Sprintf("修改任务导出状态失败 %v", sqLiteUpdateTaskRecordsErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	go ExportCSV(dataVal.TaskID, total)
	tool.Session(httpMsg, "")
}

// ExportTaskDetailByUserId 根据任务 id导出任务详情-用户
func ExportTaskDetailByUserId(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, GetExportDetailValidatorErr := validator.GetExportDetailByUserIdValidator(data)
	if GetExportDetailValidatorErr != nil {
		tool.Error(httpMsg, GetExportDetailValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	read := rep.CreateDbFactoryRead()

	//查询任务信息
	task, getTaskRecordsByTaskIdErr := read.GetTaskRecordsByTaskId(dataVal.TaskID)
	if getTaskRecordsByTaskIdErr != nil {
		errMsg := fmt.Sprintf("获取任务信息失败 %v", getTaskRecordsByTaskIdErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 验证用户
	if dataVal.UserID != fmt.Sprintf("%v", task.UserId) {
		errMsg := "用户验证失败"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	if task.IsExport == 1 {
		errMsg := "任务已导出过"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//获取任务详情总数
	total, GetBodyOverCount := service.GetBodyOverCount(dataVal.TaskID)
	if GetBodyOverCount != nil {
		errMsg := fmt.Sprintf("获取任务详情总数失败 %v", GetBodyOverCount)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//向导出任务表写入一条数据
	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	mysqlCreateTaskExportErr := mysqlWrite.CreateTaskExport(_type.TaskExportDTO{
		UserId:     task.UserId,
		ShopId:     task.ShopId,
		TaskId:     dataVal.TaskID,
		ShopName:   task.ShopName,
		FileUrl:    "",
		Status:     0,
		Total:      total,
		CompleteAt: sql.NullTime{},
	})
	if mysqlCreateTaskExportErr != nil {
		errMsg := fmt.Sprintf("写入任务信息失败 %v", mysqlCreateTaskExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	sqLiteCreateTaskExport := sqliteWrite.CreateTaskExport(_type.TaskExportDTO{
		UserId:     task.UserId,
		ShopId:     task.ShopId,
		TaskId:     dataVal.TaskID,
		ShopName:   task.ShopName,
		FileUrl:    "",
		Status:     0,
		Total:      total,
		CompleteAt: sql.NullTime{},
	})
	if sqLiteCreateTaskExport != nil {
		errMsg := fmt.Sprintf("写入任务信息失败 %v", sqLiteCreateTaskExport)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//修改任务导出状态
	mysqlUpdateTaskExportStatusErr := mysqlWrite.UpdateTaskExportStatus(dataVal.TaskID, 1, "")
	if mysqlUpdateTaskExportStatusErr != nil {
		errMsg := fmt.Sprintf("修改任务导出状态失败 %v", mysqlUpdateTaskExportStatusErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	sqLiteUpdateTaskExportStatusErr := sqliteWrite.UpdateTaskExportStatus(dataVal.TaskID, 1, "")
	if sqLiteUpdateTaskExportStatusErr != nil {
		errMsg := fmt.Sprintf("修改任务导出状态失败 %v", sqLiteUpdateTaskExportStatusErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	go ExportCSV(dataVal.TaskID, total)
	tool.Session(httpMsg, "")
}

// ExportCSV 导出CSV
// taskId 任务id
// total 总数
func ExportCSV(taskId string, total int64) {
	// 定义每次获取的数量和 CSV文件名
	batchSize := 1000
	csvFileName := fmt.Sprintf("%v.csv", taskId)

	// 初始化偏移量
	offset := 0
	// 标记是否是第一次写入（用于写入CSV表头）
	isFirstWrite := true
	// 定义导出目录
	exportDir := "export"
	// 检查并创建目录（如果不存在）
	err := os.MkdirAll(exportDir, 0755) // 0755是目录权限，保证可读可写
	if err != nil {
		errMsg := fmt.Sprintf("创建目录失败: %v", err)
		fmt.Println(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	// 拼接完整的文件路径（自动处理路径分隔符，兼容Windows/Linux）
	fullPath := filepath.Join(exportDir, csvFileName)

	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	// 更新任务导出状态-导出中
	mysqlUpdateTaskExportStatusErr := mysqlWrite.UpdateTaskExportStatus(taskId, 1, "")
	if mysqlUpdateTaskExportStatusErr != nil {
		errMsg := fmt.Sprintf("更新任务导出状态失败: %v", mysqlUpdateTaskExportStatusErr)
		fmt.Println(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}
	sqLiteUpdateTaskExportStatusErr := sqliteWrite.UpdateTaskExportStatus(taskId, 1, "")
	if sqLiteUpdateTaskExportStatusErr != nil {
		errMsg := fmt.Sprintf("更新任务导出状态失败: %v", sqLiteUpdateTaskExportStatusErr)
		fmt.Println(errMsg)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	}

	// 循环获取并写入数据
	for {
		// 每次获取1000条数据
		dataBatch, err := service.GetBodyOverDataByBatch(taskId, offset, batchSize)
		if err != nil {
			errMsg := fmt.Sprintf("获取任务详情批次数据失败 offset:%d, err:%v", offset, err)
			fmt.Println(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			return
		}

		// 没有数据了，退出循环
		if len(dataBatch) == 0 {
			//导出完成

			mysqlUpdateTaskExportStatusErr = mysqlWrite.UpdateTaskExportStatus(taskId, 2, fullPath)
			if mysqlUpdateTaskExportStatusErr != nil {
				errMsg := fmt.Sprintf("更新任务导出状态失败: %v", mysqlUpdateTaskExportStatusErr)
				fmt.Println(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			}
			sqLiteUpdateTaskExportStatusErr = sqliteWrite.UpdateTaskExportStatus(taskId, 2, fullPath)
			if sqLiteUpdateTaskExportStatusErr != nil {
				errMsg := fmt.Sprintf("更新任务导出状态失败: %v", sqLiteUpdateTaskExportStatusErr)
				fmt.Println(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			}
			//清空body_over
			clearBodyOverErr := service.ClearBodyOver(taskId)
			if clearBodyOverErr != nil {
				errMsg := fmt.Sprintf("清空body_over失败: %v", clearBodyOverErr)
				fmt.Println(errMsg)
				logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
				return
			}
			break
		}

		// 更新进度（可选）
		completed := offset
		if completed > int(total) {
			completed = int(total)
		}
		// 追加写入 CSV文件
		if writeErr := AppendToCSV(fullPath, dataBatch, isFirstWrite, taskId, &completed); writeErr != nil {
			errMsg := fmt.Sprintf("写入CSV文件失败 offset:%d, err:%v", offset, writeErr)
			fmt.Println(errMsg)
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			return
		}

		// 第一次写入后标记为false（后续不再写表头）
		if isFirstWrite {
			isFirstWrite = false
		}

		// 更新偏移量
		offset += batchSize
	}
}
