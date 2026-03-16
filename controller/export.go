package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"planA/modules/logs"
	"planA/tool"
	"time"

	"planA/service"
	sqLiteType "planA/type/sqLite"

	"github.com/gorilla/mux"
)

// GetExportTask 导出任务列表
func GetExportTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 获取分页参数
	page, size := tool.SetPage(data.URL.Query().Get("page"), data.URL.Query().Get("size"))

	records, total, getTaskExportsWithPageErr := service.GetTaskExportsWithPage(page, size, "")
	if getTaskExportsWithPageErr != nil {
		errMsg := getTaskExportsWithPageErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	dataTaskAll := []map[string]interface{}{}
	for _, v := range records {
		complete, getExportFileProgressErr := service.GetExportFileProgress(v.TaskID)
		if getExportFileProgressErr != nil {
			errMsg := "获取任务进度失败: " + getExportFileProgressErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			continue
		}
		taskExportdata := map[string]interface{}{
			"task_id":     v.TaskID,
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

// GetExportTaskUser 导出任务列表-用户
func GetExportTaskUser(httpMsg http.ResponseWriter, data *http.Request) {
	// 从路径参数获取 id
	vars := mux.Vars(data)
	userId := vars["userId"]

	// 获取分页参数
	page, size := tool.SetPage(data.URL.Query().Get("page"), data.URL.Query().Get("size"))

	records, total, getTaskExportsWithPageErr := service.GetTaskExportsWithPage(page, size, userId)
	if getTaskExportsWithPageErr != nil {
		errMsg := getTaskExportsWithPageErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	dataTaskAll := []map[string]interface{}{}
	for _, v := range records {
		complete, getExportFileProgressErr := service.GetExportFileProgress(v.TaskID)
		if getExportFileProgressErr != nil {
			errMsg := "获取任务进度失败: " + getExportFileProgressErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		taskExportdata := map[string]interface{}{
			"task_id":     v.TaskID,
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

	// 从路径参数获取 id
	vars := mux.Vars(data)
	taskId := vars["id"]

	task, getTaskRecordByTaskIDErr := service.GetTaskRecordByTaskID(taskId)

	//查询任务信息
	if getTaskRecordByTaskIDErr != nil {
		errMsg := fmt.Sprintf("获取任务信息失败 %v", getTaskRecordByTaskIDErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//查询是任务状态
	taskRecord, getTaskRecordsByTaskIDErr := service.GetTaskRecordsByTaskID(taskId)
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
	total, GetBodyOverCount := service.GetBodyOverCount(taskId)
	if GetBodyOverCount != nil {
		errMsg := fmt.Sprintf("获取任务详情总数失败 %v", GetBodyOverCount)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//写入 sqlite
	exportId, insertTaskExportErr := service.InsertTaskExport(sqLiteType.TaskExport{
		UserID:   task.UserID,
		TaskID:   taskId,
		ShopName: task.ShopName,
		FileUrl:  "",
		Status:   0,
		Total:    int(total),
		CreateAt: time.Time{},
	})
	if insertTaskExportErr != nil {
		errMsg := fmt.Sprintf("写入任务信息失败 %v", insertTaskExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//修改任务导出状态
	updateTaskRecordIsExportErr := service.UpdateTaskRecordIsExport(taskId)
	if updateTaskRecordIsExportErr != nil {
		errMsg := fmt.Sprintf("修改任务导出状态失败 %v", updateTaskRecordIsExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	UpdateTaskUserIsExportErr := service.UpdateTaskUserIsExport(taskId, 1)
	if UpdateTaskUserIsExportErr != nil {
		errMsg := fmt.Sprintf("修改任务导出状态失败 %v", UpdateTaskUserIsExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	go ExportCSV(exportId, taskId, total)
	tool.Session(httpMsg, "")
}

// ExportTaskDetailByUserId 根据任务 id导出任务详情-用户
func ExportTaskDetailByUserId(httpMsg http.ResponseWriter, data *http.Request) {

	// 从路径参数获取 id
	vars := mux.Vars(data)
	userId := vars["userId"]
	taskId := vars["id"]

	task, getTaskRecordByTaskIDErr := service.GetTaskRecordByTaskID(taskId)

	// 验证用户
	if userId != fmt.Sprintf("%v", task.UserID) {
		errMsg := "用户验证失败"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//查询任务信息
	if getTaskRecordByTaskIDErr != nil {
		errMsg := fmt.Sprintf("获取任务信息失败 %v", getTaskRecordByTaskIDErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//查询是任务状态
	taskRecord, getTaskRecordsByTaskIDErr := service.GetTaskUserByTaskId(taskId)
	if getTaskRecordsByTaskIDErr != nil {
		errMsg := fmt.Sprintf("获取任务信息失败 %v", getTaskRecordsByTaskIDErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	if *taskRecord.IsExport == 1 {
		errMsg := "任务已导出过"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//获取任务详情总数
	total, GetBodyOverCount := service.GetBodyOverCount(taskId)
	if GetBodyOverCount != nil {
		errMsg := fmt.Sprintf("获取任务详情总数失败 %v", GetBodyOverCount)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//写入 sqlite
	exportId, insertTaskExportErr := service.InsertTaskExport(sqLiteType.TaskExport{
		UserID:   task.UserID,
		TaskID:   taskId,
		ShopName: task.ShopName,
		FileUrl:  "",
		Status:   0,
		Total:    int(total),
		CreateAt: time.Time{},
	})
	if insertTaskExportErr != nil {
		errMsg := fmt.Sprintf("写入任务信息失败 %v", insertTaskExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//修改任务导出状态
	updateTaskRecordIsExportErr := service.UpdateTaskRecordIsExport(taskId)
	if updateTaskRecordIsExportErr != nil {
		errMsg := fmt.Sprintf("修改任务导出状态失败 %v", updateTaskRecordIsExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	UpdateTaskUserIsExportErr := service.UpdateTaskUserIsExport(taskId, 1)
	if UpdateTaskUserIsExportErr != nil {
		errMsg := fmt.Sprintf("修改任务导出状态失败 %v", UpdateTaskUserIsExportErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	go ExportCSV(exportId, taskId, total)
	tool.Session(httpMsg, "")
}

// ExportCSV 导出CSV
func ExportCSV(exportId int64, taskId string, total int64) {
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

	// 更新任务导出状态-导出中
	updateTaskExportStatusErr := service.UpdateTaskExportStatus(exportId, 1, "")
	if updateTaskExportStatusErr != nil {
		errMsg := fmt.Sprintf("更新任务导出状态失败: %v", updateTaskExportStatusErr)
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
			updateTaskExportStatusErr := service.UpdateTaskExportStatus(exportId, 2, fullPath)
			if updateTaskExportStatusErr != nil {
				errMsg := fmt.Sprintf("更新任务导出状态失败: %v", updateTaskExportStatusErr)
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
