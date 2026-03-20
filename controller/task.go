package controller

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"planA/rep"
	"planA/tool/process"
	"planA/validator"

	"fmt"
	"io"
	"net/http"
	"os"
	"planA/controlState/lock"
	"planA/initialization/config"
	"planA/modules/logs"
	"planA/modules/pdd"
	"planA/service"
	"planA/tool"
	"planA/type"
	redisType "planA/type/redis"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	_redis "github.com/go-redis/redis/v8"
)

// CreateTask 创建任务
func CreateTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, createTaskValidatorErr := validator.CreateTaskValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	imgType, err := strconv.ParseInt(dataVal.ImgType, 10, 64)
	if err != nil {
		errMsg := "图片类型转换失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//将 taskTypeStr 转为 int64
	taskType, err := strconv.ParseInt(dataVal.TaskType, 10, 64)
	if err != nil {
		errMsg := "任务类型转换失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 查询店铺数据
	shopDataStr, err := service.GetTaskShop(dataVal.ShopID)
	if err != nil {
		errMsg := "获取店铺数据失败: shopId " + dataVal.ShopID + " " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 解析 json数据
	shopData, err := parseShopData(shopDataStr)
	if err != nil {
		errMsg := "解析店铺数据失败:" + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	shop := shopData.Shop
	spec := shopData.Spec
	detail := shopData.ShopDetail
	context := shopData.ShopContext
	priceTemplateRangeStr := shopData.PriceTemplate.RangePrice

	var priceRange []_type.PriceRange
	err = json.Unmarshal([]byte(priceTemplateRangeStr), &priceRange)
	if err != nil {
		errMsg := "解析价格模板失败:" + err.Error() + " 原始数据：" + shopData.PriceTemplate.RangePrice
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	if shop.ShopType != dataVal.ShopType {
		errMsg := "店铺类型不匹配 错误店铺类型:" + shop.ShopType
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//验证店铺规格信息是否正确
	if dataVal.ShopType == "1" {
		pddDll, initPddSOErr := pdd.InitPddDll()
		if initPddSOErr != nil {
			errMsg := "初始化pdd.so失败: " + initPddSOErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		_, buildPddGoodsSpecIdErr := buildPddGoodsSpecId(pddDll, shop.Token, spec.SpecTypeID, spec.SpecName)
		if buildPddGoodsSpecIdErr != nil {
			errMsg := "构建规格ID失败: " + buildPddGoodsSpecIdErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	}

	//扣费
	//userId := strconv.FormatInt(shop.CreateBy, 10)
	//_, taskDeductionErr := TaskDeduction(shopID, userId)
	//if taskDeductionErr != nil {
	//	errMsg := "请求创建任务接口失败: " + taskDeductionErr.Error() + "店铺id：" + shopID + "用户id：" + userId
	//	tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
	//	return
	//}

	//请求创建任务接口并获取任务 id
	taskId, err := CreateTaskRequest(dataVal.ShopID)
	if err != nil {
		errMsg := "请求创建任务接口失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	fmt.Printf("店铺ID: %s, 店铺类型: %s, 任务类型: %s, 任务数量: %s 任务id: %s \n", dataVal.ShopID, dataVal.ShopType, dataVal.TaskType, dataVal.TaskCount, taskId)

	// 创建任务逻辑...
	createAt := time.Now().Unix()
	task, err := CreateTaskData(taskId, taskType, createAt, shop, priceRange, spec, detail, context, dataVal.TaskCount, imgType)
	if err != nil {
		errMsg := "创建任务失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//推送redis
	err = service.UpdateTaskHeader(taskId, task.Header)
	if err != nil {
		errMsg := "保存任务头失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	err = service.UpdateTaskFooter(taskId, &task.Footer)
	if err != nil {
		errMsg := "保存任务尾失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()

	//写入 mysql数据
	mysqlCreateTaskRecordsErr := mysqlWrite.CreateTaskRecords(_type.TaskRecordsDTO{
		UserId:   strconv.FormatInt(shopData.Shop.CreateBy, 10),
		ShopId:   strconv.FormatInt(shopData.Shop.ID, 10),
		TaskId:   taskId,
		ShopName: shop.ShopName,
		TaskType: taskType,
	})
	if mysqlCreateTaskRecordsErr != nil {
		errMsg := "插入任务用户失败: " + mysqlCreateTaskRecordsErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//写入 sqlite数据
	sqliteTaskExportErr := sqliteWrite.CreateTaskRecords(_type.TaskRecordsDTO{
		UserId:   strconv.FormatInt(shopData.Shop.CreateBy, 10),
		ShopId:   strconv.FormatInt(shopData.Shop.ID, 10),
		TaskId:   taskId,
		ShopName: shop.ShopName,
		TaskType: taskType,
	})
	if sqliteTaskExportErr != nil {
		errMsg := "插入任务用户失败: " + sqliteTaskExportErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	tool.Session(httpMsg, taskId)
}

// SetTaskBody 置任务体
func SetTaskBody(httpMsg http.ResponseWriter, data *http.Request) {

	// 方法1：直接使用multipart reader（最安全）
	contentType := data.Header.Get("Content-Type")
	if !strings.Contains(contentType, "multipart/form-data") {
		tool.Error(httpMsg, "Content-Type必须是multipart/form-data", http.StatusBadRequest)
		return
	}

	// 移除请求体大小限制
	const maxInt64 = 1<<63 - 1
	data.Body = http.MaxBytesReader(httpMsg, data.Body, maxInt64)

	// 创建multipart reader
	reader, err := data.MultipartReader()
	if err != nil {
		tool.Error(httpMsg, "创建multipart reader失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var bodyData []string
	var taskId string

	// 流式处理每个部分
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			tool.Error(httpMsg, "读取表单部分失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 读取这部分的内容
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, part); err != nil {
			tool.Error(httpMsg, "读取数据失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		content := buf.String()
		formName := part.FormName()

		if formName == "body" {
			bodyData = append(bodyData, content)
		} else if formName == "task_id" {
			taskId = content
		}
	}
	// 验证任务 ID
	if taskId == "" {
		errMsg := "任务 ID 不能为空"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//验证状态
	header, getTaskHeaderErr := service.GetTaskHeader(taskId)
	if getTaskHeaderErr != nil {
		errMsg := "获取任务头失败: " + getTaskHeaderErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	if header.Status == _type.TaskStatusStopped {
		errMsg := "任务已停止"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 更新任务数
	go UpdateTaskCount(bodyData, taskId)

	// 返回成功响应
	tool.Session(httpMsg, "")
}

// PauseTask 暂停任务
func PauseTask(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.UpdateTaskStatusValidator(data)
	if updateTaskStatusValidatorErr != nil {
		tool.Error(httpMsg, updateTaskStatusValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	// 验证状态
	header, getTaskHeaderErr := service.GetTaskHeader(dataVal.TaskID)
	if getTaskHeaderErr != nil {
		errMsg := "获取任务头失败: " + getTaskHeaderErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	if header.Status != _type.TaskStatusRunning {
		errMsg := "当前状态不是执行中"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	read := rep.CreateDbFactoryRead()
	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	// 查询当前任务信息
	taskRecords, getTaskRecordsByTaskIdErr := read.GetTaskRecordsByTaskId(dataVal.TaskID)
	if getTaskRecordsByTaskIdErr != nil {
		errMsg := fmt.Sprintf("获取任务信息失败 %v", getTaskRecordsByTaskIdErr)
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 查询当前导出任务信息
	taskExport, getTaskExportByTaskIdErr := read.GetTaskExportByTaskId(dataVal.TaskID)
	if getTaskExportByTaskIdErr != nil {
		return
	}
	// 暂停时将task_records表状态改为未导出状态
	mysqlUpdateTaskRecordsErr := mysqlWrite.UpdateTaskRecords(_type.TaskRecordsDTO{
		UserId:   taskRecords.UserId,
		ShopId:   taskRecords.ShopId,
		TaskId:   taskRecords.TaskId,
		ShopName: taskRecords.ShopName,
		IsExport: 0,
		TaskType: taskRecords.TaskType,
	})
	if mysqlUpdateTaskRecordsErr != nil {
		errMsg := "更新任务用户失败: " + mysqlUpdateTaskRecordsErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	sqliteUpdateTaskRecordsErr := sqliteWrite.UpdateTaskRecords(_type.TaskRecordsDTO{
		UserId:   taskRecords.UserId,
		ShopId:   taskRecords.ShopId,
		TaskId:   taskRecords.TaskId,
		ShopName: taskRecords.ShopName,
		IsExport: 0,
		TaskType: taskRecords.TaskType,
	})
	if sqliteUpdateTaskRecordsErr != nil {
		errMsg := "更新任务用户失败: " + sqliteUpdateTaskRecordsErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 暂停时将task_export状态改为未导出状态
	mysqlUpdateTaskExportStatusErr := mysqlWrite.UpdateTaskExportStatus(taskExport.TaskId, 1, taskExport.FileUrl)
	if mysqlUpdateTaskExportStatusErr != nil {
		errMsg := "更新任务用户失败: " + mysqlUpdateTaskExportStatusErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	sqliteUpdateTaskExportStatusErr := sqliteWrite.UpdateTaskExportStatus(taskExport.TaskId, 1, taskExport.FileUrl)
	if sqliteUpdateTaskExportStatusErr != nil {
		errMsg := "更新任务用户失败: " + sqliteUpdateTaskExportStatusErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 暂停 B程序
	suspendProcessErr := process.SuspendProcess(dataVal.TaskID)
	if suspendProcessErr != nil {
		errMsg := "暂停任务失败: " + suspendProcessErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 返回成功响应
	tool.Session(httpMsg, "")
}

// ResumeTask 恢复任务
func ResumeTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.UpdateTaskStatusValidator(data)
	if updateTaskStatusValidatorErr != nil {
		tool.Error(httpMsg, updateTaskStatusValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	//验证状态
	header, getTaskHeaderErr := service.GetTaskHeader(dataVal.TaskID)
	if getTaskHeaderErr != nil {
		errMsg := "获取任务头失败: " + getTaskHeaderErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	if header.Status != _type.TaskStatusPaused {
		errMsg := "当前状态不是暂停"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//推送redis
	status := int64(_type.TaskStatusRunning)
	err := service.UpdateHeaderStatus(dataVal.TaskID, status)
	if err != nil {
		errMsg := err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 恢复 B程序
	suspendProcessErr := process.ResumeProcess(dataVal.TaskID)
	if suspendProcessErr != nil {
		errMsg := "恢复进程失败: " + suspendProcessErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 返回成功响应
	tool.Session(httpMsg, "")
}

// StopTask 停止任务
func StopTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.UpdateTaskStatusValidator(data)
	if updateTaskStatusValidatorErr != nil {
		tool.Error(httpMsg, updateTaskStatusValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	//验证状态
	header, getTaskHeaderErr := service.GetTaskHeader(dataVal.TaskID)
	if getTaskHeaderErr != nil {
		errMsg := "获取任务头失败: " + getTaskHeaderErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	if header.Status == _type.TaskStatusOver {
		errMsg := "任务已完成"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 停止 B程序
	stopProcessErr := process.StopTask(dataVal.TaskID)
	if stopProcessErr != nil {
		errMsg := "停止进程失败: " + stopProcessErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	tool.Session(httpMsg, "")
}

// DelTask 删除任务
func DelTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.UpdateTaskStatusValidator(data)
	if updateTaskStatusValidatorErr != nil {
		tool.Error(httpMsg, updateTaskStatusValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	//获取任务状态
	header, getTaskHeaderErr := service.GetTaskHeader(dataVal.TaskID)
	if getTaskHeaderErr != nil {
		errMsg := "获取任务头失败: " + getTaskHeaderErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 如果任务是暂停则先恢复
	if header.Status == _type.TaskStatusPaused {
		// 恢复 B程序
		suspendProcessErr := process.ResumeProcess(dataVal.TaskID)
		if suspendProcessErr != nil {
			errMsg := "恢复进程失败: " + suspendProcessErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}

		// 停止 B程序 清空任务
		stopProcessErr := process.StopTask(dataVal.TaskID)
		if stopProcessErr != nil {
			errMsg := "停止进程失败: " + stopProcessErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	}

	// 删除 redis中的内容
	mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
	delTaskErr := service.DelTask(dataVal.TaskID)
	if delTaskErr != nil {
		errMsg := "删除任务失败: " + delTaskErr.Error()
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
	//删除 mysql中TaskRecords指定数据
	mysqlDeleteTaskRecordsByTaskIdErr := mysqlWrite.DeleteTaskRecordsByTaskId(dataVal.TaskID)
	if mysqlDeleteTaskRecordsByTaskIdErr != nil {
		errMsg := "删除任务失败: " + delTaskErr.Error()
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
	// 删除 sqlite中TaskRecords指定数据
	sqLiteDeleteTaskRecordsByTaskIDErr := sqliteWrite.DeleteTaskRecordsByTaskId(dataVal.TaskID)
	if sqLiteDeleteTaskRecordsByTaskIDErr != nil {
		errMsg := "删除任务失败: " + delTaskErr.Error()
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return
	}
	//3秒后再次删除，避免删除期间body进入数据
	go func() {
		mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()
		// 删除任务 延迟3后删除
		time.Sleep(time.Duration(3) * time.Second)
		delTaskErr := service.DelTask(dataVal.TaskID)
		if delTaskErr != nil {
			errMsg := "删除任务失败: " + delTaskErr.Error()
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			return
		}
		//删除 mysql中TaskRecords指定数据
		mysqlDeleteTaskRecordsByTaskIdErr := mysqlWrite.DeleteTaskRecordsByTaskId(dataVal.TaskID)
		if mysqlDeleteTaskRecordsByTaskIdErr != nil {
			errMsg := "删除任务失败: " + delTaskErr.Error()
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			return
		}
		// 删除 sqlite中TaskRecords指定数据
		sqLiteDeleteTaskRecordsByTaskIDErr := sqliteWrite.DeleteTaskRecordsByTaskId(dataVal.TaskID)
		if sqLiteDeleteTaskRecordsByTaskIDErr != nil {
			errMsg := "删除任务失败: " + delTaskErr.Error()
			logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
			return
		}
	}()

	tool.Session(httpMsg, "")
}

// OverTask 任务完成
func OverTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.UpdateTaskStatusValidator(data)
	if updateTaskStatusValidatorErr != nil {
		tool.Error(httpMsg, updateTaskStatusValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	//推送 redis
	status := int64(_type.TaskStatusOver)
	err := service.UpdateHeaderStatus(dataVal.TaskID, status)
	if err != nil {
		errMsg := err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	lock.DestroyLock(dataVal.TaskID) //销毁锁
	// 返回成功响应
	tool.Session(httpMsg, "")
}

// GetTask 任务列表
func GetTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, getTaskValidatorErr := validator.GetTaskValidator(data)
	if getTaskValidatorErr != nil {
		tool.Error(httpMsg, getTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)

	taskTypeInt64 := int64(0)
	var taskTypeAtoiErr error
	if dataVal.TaskType != "" {
		//将 taskTypeStr 转为 int
		var temp int
		temp, taskTypeAtoiErr = strconv.Atoi(dataVal.TaskType)
		if taskTypeAtoiErr != nil {
			errMsg := "任务类型转换失败"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		taskTypeInt64 = int64(temp)
	}

	read := rep.CreateDbFactoryRead()
	records, total, getTaskRecordsListErr := read.GetTaskRecordsList(_type.GetTaskRecordsListReq{
		UserId:   "",
		TaskId:   dataVal.TaskID,
		TaskType: taskTypeInt64,
		ShopName: dataVal.ShopName,
		Page:     page,
		Size:     size,
	})
	if getTaskRecordsListErr != nil {
		errMsg := getTaskRecordsListErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	dataTaskAll := []map[string]interface{}{}
	for _, v := range records {
		//查询 header 信息
		header, getTaskHeaderErr := service.GetTaskHeader(v.TaskId)
		if getTaskHeaderErr != nil {
			errMsg := fmt.Sprintf("获取footer 信息失败 %v", getTaskHeaderErr)
			fmt.Println(errMsg)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//获取 footer 信息
		footer, getTaskFooterErr := service.GetTaskFooter(v.TaskId)
		if getTaskFooterErr != nil {
			errMsg := fmt.Sprintf("获取footer 信息失败 %v", getTaskFooterErr)
			fmt.Println(errMsg)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//获取 body_over 信息
		bodyOver, GetTaskBodyOverLimit10Err := service.GetTaskBodyOverLimit10(v.TaskId)
		if GetTaskBodyOverLimit10Err != nil {
			errMsg := fmt.Sprintf("获取body_over 信息失败 %v", GetTaskBodyOverLimit10Err)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		footerData := map[string]interface{}{
			"task_count_true":    footer.TaskCountTrue,
			"task_count_success": footer.TaskCountSuccess.Load(),
			"task_count_error":   footer.TaskCountError.Load(),
			"task_count_wait":    footer.TaskCountWait.Load(),
			"task_count_over":    footer.TaskCountOver.Load(),
			"task_qpm":           footer.TaskQpm,
			"last_index":         footer.LastIndex,
			"task_count":         footer.TaskCount,
		}
		dataTask := map[string]interface{}{
			"header":    header,
			"footer":    footerData,
			"body_over": bodyOver,
			"is_export": v.IsExport,
		}
		dataTaskAll = append(dataTaskAll, dataTask)
	}
	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  dataTaskAll,
	}
	tool.Session(httpMsg, dataRet)
}

// GetTaskByUserId 获取用户任务
func GetTaskByUserId(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, getTaskByUserIdValidatorErr := validator.GetTaskByUserIdValidator(data)
	if getTaskByUserIdValidatorErr != nil {
		tool.Error(httpMsg, getTaskByUserIdValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	// 获取分页参数
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	taskTypeInt64 := int64(0)
	var parseIntTaskTypeErr error
	if dataVal.TaskType != "" {
		//将taskType 转换为 int64
		taskTypeInt64, parseIntTaskTypeErr = strconv.ParseInt(dataVal.TaskType, 10, 64)
		if parseIntTaskTypeErr != nil {
			errMsg := "任务类型转换失败"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	}

	read := rep.CreateDbFactoryRead()
	records, total, GetTaskUserListErr := read.GetTaskRecordsList(_type.GetTaskRecordsListReq{
		UserId:   dataVal.UserID,
		TaskId:   dataVal.TaskID,
		TaskType: taskTypeInt64,
		ShopName: dataVal.ShopName,
		Page:     page,
		Size:     size,
	})
	if GetTaskUserListErr != nil {
		errMsg := GetTaskUserListErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	dataTaskAll := []map[string]interface{}{}
	for _, v := range records {
		//查询 header 信息
		header, getTaskHeaderErr := service.GetTaskHeader(v.TaskId)
		if getTaskHeaderErr != nil {
			errMsg := fmt.Sprintf("获取footer 信息失败 %v", getTaskHeaderErr)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//获取 footer 信息
		footer, getTaskFooterErr := service.GetTaskFooter(v.TaskId)
		if getTaskFooterErr != nil {
			errMsg := fmt.Sprintf("获取footer 信息失败 %v", getTaskFooterErr)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//获取 body_over 信息
		bodyOver, GetTaskBodyOverLimit10Err := service.GetTaskBodyOverLimit10(v.TaskId)
		if GetTaskBodyOverLimit10Err != nil {
			errMsg := fmt.Sprintf("获取body_over 信息失败 %v", getTaskFooterErr)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		footerData := map[string]interface{}{
			"task_count_true":    footer.TaskCountTrue,
			"task_count_success": footer.TaskCountSuccess.Load(),
			"task_count_error":   footer.TaskCountError.Load(),
			"task_count_wait":    footer.TaskCountWait.Load(),
			"task_count_over":    footer.TaskCountOver.Load(),
			"task_qpm":           footer.TaskQpm,
			"last_index":         footer.LastIndex,
			"task_count":         footer.TaskCount,
		}
		dataTask := map[string]interface{}{
			"header":    header,
			"footer":    footerData,
			"body_over": bodyOver,
			"is_export": v.IsExport,
		}
		dataTaskAll = append(dataTaskAll, dataTask)
	}
	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  dataTaskAll,
	}
	tool.Session(httpMsg, dataRet)
}

// GetTaskHeader 获取 header信息
func GetTaskHeader(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, getHeaderValidatorErr := validator.GetHeaderValidator(data)
	if getHeaderValidatorErr != nil {
		tool.Error(httpMsg, getHeaderValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	header, getTaskHeaderErr := service.GetTaskHeader(dataVal.TaskID)
	if getTaskHeaderErr != nil {
		errMsg := getTaskHeaderErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	tool.Session(httpMsg, header)
}

func B(httpMsg http.ResponseWriter, data *http.Request) {
	taskID := "111"
	_, callSendPublishingErr := process.RunTaskWorker(taskID)
	if callSendPublishingErr != nil {
		logStr := fmt.Sprintf("执行B程序失败: [taskId] %v [error] %v", taskID, callSendPublishingErr.Error())
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, logStr)
		tool.Error(httpMsg, callSendPublishingErr.Error(), http.StatusInternalServerError)
		return
	}
	tool.Session(httpMsg, "")
}

//****************************工具**************************************//

// parseShopData 解析店铺数据
// @param shopData 店铺数据
// @return *_type.ShopInfo 店铺信息
func parseShopData(shopData string) (*_type.ShopInfo, error) {
	shopData = strings.TrimSpace(shopData)

	// 直接解析为 RedisData数组
	var redisData []redisType.RedisData
	err := json.Unmarshal([]byte(shopData), &redisData)
	if err != nil {
		// 尝试另一种格式：可能是单对象而不是数组
		var singleData redisType.RedisData
		if singleErr := json.Unmarshal([]byte(shopData), &singleData); singleErr == nil {
			redisData = []redisType.RedisData{singleData}
		} else {
			return nil, fmt.Errorf("JSON解析失败: %v, 原始数据: %s", err, shopData[:min(100, len(shopData))])
		}
	}

	shopInfo := &_type.ShopInfo{}

	// 遍历所有数据，根据source_table分类
	for _, item := range redisData {
		switch item.SourceTable {
		case "t_shop":
			var shop _type.Shop
			if err := json.Unmarshal(item.Data, &shop); err == nil {
				shopInfo.Shop = &shop
			} else {
				fmt.Printf("解析t_shop失败: %v\n", err)
			}
		case "t_shop_detail":
			var detail _type.ShopDetail
			if err := json.Unmarshal(item.Data, &detail); err == nil {
				shopInfo.ShopDetail = &detail
			} else {
				fmt.Printf("解析t_shop_detail失败: %v\n", err)
			}
		case "t_shop_context":
			var context _type.ShopContext
			if err := json.Unmarshal(item.Data, &context); err == nil {
				shopInfo.ShopContext = &context
			} else {
				fmt.Printf("解析t_shop_context失败: %v\n", err)
			}
		case "t_spec":
			var spec _type.Spec
			if err := json.Unmarshal(item.Data, &spec); err == nil {
				shopInfo.Spec = &spec
			} else {
				fmt.Printf("解析t_spec失败: %v\n", err)
			}
		case "t_price_template":
			var template _type.PriceTemplate
			if err := json.Unmarshal(item.Data, &template); err == nil {
				shopInfo.PriceTemplate = &template
			} else {
				fmt.Printf("解析t_price_template失败: %v\n", err)
			}
		default:
			fmt.Printf("未知的source_table: %s\n", item.SourceTable)
		}
	}

	return shopInfo, nil
}

// CreateTaskData 创建task数据
// @param taskId 任务ID
// @param taskType 任务类型
// @param createAt 创建时间
// @param shop 店铺信息
// @param priceRange 价格模版
// @param spec 商品规格
// @param context 店铺描述
// @param taskCount 任务数量
// @param detail 店铺详情
// @param imgType 图片类型
// @return *_type.Task 任务数据
// @return error 错误
func CreateTaskData(taskId string, taskType int64, createAt int64, shop *_type.Shop, priceRange []_type.PriceRange, spec *_type.Spec, detail *_type.ShopDetail, context *_type.ShopContext, taskCount string, imgType int64) (*_type.Task, error) {
	var task _type.Task
	//处理价格模版
	var priceModArr []_type.PriceMod
	for _, v := range priceRange {

		adjustPercentInt64, err := parseAdjustPercent(v.AdjustPercent)
		if err != nil {
			return &task, fmt.Errorf("价格模版 adjustPercent 转换失败: %v", err)
		}
		priceMod := _type.PriceMod{
			Min:         v.MinPrice,
			Max:         v.MaxPrice,
			MarkupRate:  adjustPercentInt64,
			MarkupValue: v.AdjustAmount,
		}
		priceModArr = append(priceModArr, priceMod)
	}
	var goodsDetailFirstImgUrlArrayToDo []string
	var carouseLastImgUrlArrayToDo []string
	var goodsDetailLastImgUrlArrayToDo []string
	var token string
	var districtId int64
	var districtType string
	//处理 Token
	if shop.ShopType == "1" { //拼多店铺
		token = shop.Token
	} else if shop.ShopType == "5" { // 闲鱼店铺
		token = fmt.Sprintf("{\"app_id\":%v,\"app_secret\":\"%v\",\"username\":\"%v\"}", shop.MallID, shop.Token, shop.ShopKey)
		districtId = detail.DistrictId
		districtType = detail.DistrictType
	}
	// specTypeID 转换为int64
	specTypeID, err := parseAdjustPercent(spec.SpecTypeID)
	if err != nil {
		return &task, fmt.Errorf("价格模版 adjustPercent 转换失败: %v", err)
	}
	//shopCount 转换为int64
	taskCountInt64, err := strconv.ParseInt(taskCount, 10, 64)
	if err != nil {
		return &task, fmt.Errorf("shopCount 转换为int64 转换失败: %v", err)
	}
	//发货时间
	shipmentLimitSecond := int64(24 * 60 * 60) //默认发货时间24小时
	if detail.ShipmentLimitSecond != "1" {
		shipmentLimitSecond = shipmentLimitSecond * 2 //发货时间48小时
	}
	task = _type.Task{
		Header: _type.TaskHeader{
			TaskId:   taskId,
			TaskType: taskType,
			ShopId:   shop.ID,
			ShopName: shop.ShopName,
			ShopType: shop.ShopType,
			ShopMsg: _type.ShopMsg{
				ID:                          detail.ID,                                           //店铺详情 ID
				ShopAliasName:               shop.ShopName,                                       //店铺别名
				ShopName:                    shop.ShopName,                                       //店铺名称
				Token:                       token,                                               //店铺 token【如果是咸鱼店铺，此token则是应用密钥】
				GoodsNamePrefix:             detail.TitlePrefix,                                  //商品名称前缀
				GoodsNameSuffix:             detail.TitleSuffix,                                  //商品名称后缀
				TitleConsistOf:              detail.TitleConsistOf,                               //商品名称组成
				SpaceCharacter:              detail.SpaceCharacter,                               //间隔字符  0无间隔 1空格
				WatermarkImgUrl:             detail.WatermarkImgUrl,                              //水印图片
				WatermarkPosition:           detail.WatermarkPosition,                            //水印位置 0全部  1第一张
				CarouseLastImgUrlArray:      tool.FilterStrings(carouseLastImgUrlArrayToDo),      //轮播图最后图片[]string（tool.FilterStrings 函数为去掉数组中的空、图片不合法等字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				GoodsDetailFirstImgUrlArray: tool.FilterStrings(goodsDetailFirstImgUrlArrayToDo), //商品详情首图URL数组[]string（tool.FilterStrings 函数为去掉数组中的空、图片不合法字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				GoodsDetailLastImgUrlArray:  tool.FilterStrings(goodsDetailLastImgUrlArrayToDo),  //商品详情最后图片URL数组（tool.FilterStrings 函数为去掉数组中的空、图片不合法等字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				IsFolt:                      detail.Fake == "1",                                  //是否支持假一赔十，false-不支持，true-支持
				IsPreSale:                   detail.Presale == "1" || detail.Presale == "2",      //是否预售,true-预售商品，false-非预售商品
				IsRefundable:                detail.SevenDays == "1",                             //是否7天无理由退换货，true-支持，false-不支持
				ShipmentLimitSecond:         shipmentLimitSecond,                                 //承诺发货时间（秒）
				CostTemplateId:              int64(detail.TemplateId),                            //物流运费模板 ID
				SpecName:                    spec.SpecTypeName,                                   //规格名称
				SpecId:                      specTypeID,                                          //规格 ID
				SpecChildName:               spec.SpecName,                                       //规格子名称
				DefStock:                    int64(detail.StockDeff),                             //默认库存
				TwoDiscount:                 detail.TowDiscount,                                  //2折
				IsSecondHand:                detail.IsSecondHand == "1",                          //是否二手 1 -二手商品 ，0-全新商品
				DistrictMsg: _type.DistrictMsg{
					DistrictId:   districtId,
					DistrictType: districtType,
				},
				ShopContext: context.Context, //店铺描述
			},
			PriceMod:         priceModArr,             //价格模版
			ShipPriceMod:     "",                      //运费模版
			TaskCount:        taskCountInt64,          //任务数量
			TaskCountTrue:    0,                       //真实任务数量
			TaskCountWait:    0,                       //等待任务数量
			TaskCountOver:    0,                       //任务完成数量
			TaskCountSuccess: 0,                       //任务成功数量
			TaskCountError:   0,                       //任务失败数量
			Status:           _type.TaskStatusRunning, //任务状态 1=运行中 2=暂停中 3=完成
			TaskQpm:          0,                       //任务QPM
			TaskCreateAt:     createAt,                //任务创建时间
			TaskOverAt:       0,                       //任务完成时间
			LastIndex:        0,                       //最后索引
			ImgType:          imgType,                 //图片类型 0=无图片 1=轮播图 2=商品详情首图 3=商品详情最后图片
			Pool: _type.PoolConfig{
				Size:                 0,
				WithExpiryDuration:   10,
				WithPreAlloc:         true,
				WithMaxBlockingTasks: 2000,
				WithNonblocking:      true,
			},
		},
		BodyOver: _type.TaskBody{},
		Footer: _type.TaskFooter{
			TaskCount:        taskCountInt64, //任务数量
			TaskCountTrue:    0,              //真实任务数量
			TaskCountWait:    atomic.Int64{}, //等待任务数量
			TaskCountOver:    atomic.Int64{}, //任务完成数量
			TaskCountSuccess: atomic.Int64{}, //任务成功数量
			TaskCountError:   atomic.Int64{}, //任务失败数量
			TaskQpm:          0,              //任务QPM
			LastIndex:        0,              //最后索引
		},
	}
	return &task, nil
}

// UpdateTaskCount 更新任务数量
// @param bodyData body数据
// @param taskId 任务ID
func UpdateTaskCount(bodyData []string, taskId string) {
	// 1. 先执行AddTask，统一判断是否需要后续操作
	count := AddTask(taskId, bodyData)
	if count <= 0 {
		fmt.Println("找到的书品为0，所以不提交到redis")
		return
	}
	// 执行 B方法程序
	_, runTaskWorkerErr := process.RunTaskWorker(taskId)
	if runTaskWorkerErr != nil {
		fmt.Printf("执行B程序出错: %v\n", runTaskWorkerErr)
		return
	}
}

func AddTask(taskId string, bodyData []string) int {

	//查询 header 信息
	header, getTaskHeaderErr := service.GetTaskHeader(taskId)
	if getTaskHeaderErr != nil {
		fmt.Printf("获取footer 信息失败 %v", getTaskHeaderErr)
		return 0
	}
	if header.Status == _type.TaskStatusOver {
		updateHeaderStatusErr := service.UpdateHeaderStatus(taskId, int64(_type.TaskStatusRunning))
		if updateHeaderStatusErr != nil {
			fmt.Printf("更新header 状态失败 %v", updateHeaderStatusErr)
			return 0
		}
	}
	// 遍历 bodyData 写入redis
	var num atomic.Int64
	for _, v := range bodyData {
		var taskBody _type.TaskBody
		// 清理JSON字符串（去除可能的空格和换行）
		jsonStr := strings.TrimSpace(v)
		if err := json.Unmarshal([]byte(jsonStr), &taskBody); err != nil {
			fmt.Printf("解析失败: %v\n", err)
			continue
		}
		// 连接DB[b] 获取书品信息
		bookInfo, GetTaskBookErr := service.GetTaskBook(taskBody.BookInfo.Isbn)
		if GetTaskBookErr != nil {
			if errors.Is(GetTaskBookErr, _redis.Nil) {
				setNoBookCountErr := service.SetNoBookCount(taskBody.BookInfo.Isbn)
				if setNoBookCountErr != nil {
					fmt.Printf("设置无书品数量失败 isbn:%v", taskBody.BookInfo.Isbn)
				}
			}
			fmt.Printf("获取BookInfo失败-原因: %v\n", GetTaskBookErr)
			continue
		}
		var catId string
		pinDuoDuoCatIdArr := tool.StringToArray(bookInfo.CatIdObject.PinDuoDuoCatId.String())
		if len(pinDuoDuoCatIdArr) == 3 {
			catId = pinDuoDuoCatIdArr[2]
		} else if len(pinDuoDuoCatIdArr) == 4 {
			catId = pinDuoDuoCatIdArr[3]
		}
		if header.ShopType == "1" {
			bookInfo.CatIdObject.PinDuoDuoCatId = _type.FlexibleStr(catId)
		} else if header.ShopType == "5" {
			bookInfo.CatIdObject.XianYuCatId = _type.FlexibleStr(bookInfo.CatIdObject.XianYuCatId.String())
		}
		// 更新 BookInfo
		taskBody.BookInfo = bookInfo

		// 更新 BodyWait
		err := service.UpdateTaskBodyWait(taskId, taskBody)
		if err != nil {
			fmt.Println(err.Error())
			return 0
		}
		//延迟1毫秒
		num.Add(1)
		err = service.UpdateTaskCountTrue(taskId, 1)
	}
	taskNoticeRequestErr := TaskNoticeRequest(taskId)
	if taskNoticeRequestErr != nil {
		return 0
	}
	return int(num.Load())
}

// 处理adjustPercent字段（可能是int或string）
// @param adjustPercent adjustPercent字段
// @return int64 处理后的数据
// @return error 错误信息
func parseAdjustPercent(adjustPercent interface{}) (int64, error) {
	if adjustPercent == nil {
		return 0, nil
	}
	//判断 adjustPercent 是否字符串 如果是 字符串转为 int64
	if reflect.TypeOf(adjustPercent).Kind() == reflect.String {
		adjustPercentStr := adjustPercent.(string)
		adjustPercentInt, err := strconv.Atoi(adjustPercentStr)
		if err != nil {
			return 0, err
		}
		return int64(adjustPercentInt), nil
	}
	//如果是 float64
	if reflect.TypeOf(adjustPercent).Kind() == reflect.Float64 {
		return int64(adjustPercent.(float64)), nil
	}
	return adjustPercent.(int64), nil
}

// CreateTaskRequest 请求接口创建任务
func CreateTaskRequest(shopId string) (string, error) {

	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		errMsg := "获取文件路径配置失败: " + getFileUrlConfigErr.Error()
		return "", fmt.Errorf(errMsg)
	}
	dataMap := map[string]string{
		"shopId":   shopId,
		"taskType": "NEW_ADD_TASK",
		"fileName": "新发布商品任务",
	}
	taskDataStr, submitFormDataErr := tool.SubmitFormData(fileUrlConfig.CreateTaskUrl, dataMap)
	if submitFormDataErr != nil {
		errMsg := "提交表单数据失败 " + submitFormDataErr.Error()
		return "", fmt.Errorf(errMsg)
	}
	var taskData _type.CreateTaskResponse
	unmarshalErr := json.Unmarshal([]byte(taskDataStr), &taskData)
	if unmarshalErr != nil {
		errMsg := "解析任务数据失败: " + unmarshalErr.Error() + " 原始数据" + taskDataStr
		return "", fmt.Errorf(errMsg)
	}
	if taskData.Code != 200 {
		errMsg := "请求接口 " + fileUrlConfig.CreateTaskUrl + " data  " + taskDataStr
		return "", fmt.Errorf(errMsg)
	}
	return taskData.TaskID, nil
}

// TaskNoticeRequest 任务有等待数据通知接口
func TaskNoticeRequest(taskId string) error {

	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		return fmt.Errorf("获取文件路径配置失败: %v", getFileUrlConfigErr)
	}
	data := map[string]string{
		"taskId": taskId,
	}
	_, submitFormDataErr := tool.SubmitFormData(fileUrlConfig.CreateTaskNoticeUrl, data)
	if submitFormDataErr != nil {
		return fmt.Errorf("提交表单数据失败: %v", submitFormDataErr)
	}
	return nil
}

// TaskDeduction 创建任务扣费
func TaskDeduction(shopId string, userId string) (_type.TaskDeductionResponse, error) {
	var taskDeductionData _type.TaskDeductionResponse
	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		return taskDeductionData, fmt.Errorf("获取文件路径配置失败: %v", getFileUrlConfigErr)
	}

	dataMap := map[string]string{
		"userId":       userId,
		"shopId":       shopId,
		"logType":      "2",
		"rechargPrice": "1",
	}
	taskDeductionStr, submitFormDataErr := tool.SubmitFormData(fileUrlConfig.DeductionUrl, dataMap)
	if submitFormDataErr != nil {
		return taskDeductionData, fmt.Errorf("提交表单数据失败 %v", submitFormDataErr)
	}
	unmarshalErr := json.Unmarshal([]byte(taskDeductionStr), &taskDeductionData)
	if unmarshalErr != nil {
		errMsg := "解析任务数据失败: " + unmarshalErr.Error() + " 原始数据" + taskDeductionStr
		return taskDeductionData, fmt.Errorf(errMsg)
	}
	if taskDeductionData.Code != 200 {
		errMsg := "请求接口 " + fileUrlConfig.CreateTaskUrl + " data  " + taskDeductionStr
		return taskDeductionData, fmt.Errorf(errMsg)
	}
	return taskDeductionData, nil
}

// AppendToCSV 追加写入数据到CSV文件
// @param fileName 文件名
// @param data 数据
// @param writeHeader 是否写入表头
// @param taskId 任务ID
// @return error
func AppendToCSV(fileName string, data []_type.TaskBody, writeHeader bool, taskId string) error {

	// 打开文件（不存在则创建，存在则追加）
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 第一次写入时添加表头
	if writeHeader && len(data) > 0 {
		// 根据TaskBody的字段定义表头，这里需要你根据实际结构体调整
		headers := []string{
			"ISBN", "书名", "状态", "错误信息", // 示例表头，替换为你实际的字段名
		}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("写入CSV表头失败: %v", err)
		}
	}

	// 写入数据行
	for _, item := range data {
		statusStr := "正确"
		if item.Detail.Status != 1 {
			statusStr = "错误"
		}
		// 将TaskBody转换为字符串切片，需要根据实际结构体字段调整
		record := []string{
			item.BookInfo.Isbn,
			item.BookInfo.BookName,
			statusStr,
			item.Detail.Error,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入CSV数据失败: %v, 数据: %+v", err, item)
		}
		// 更新redis中的Complete字段，展示导出进度
		err := service.UpdateExportFileProgress(taskId)
		if err != nil {
			return fmt.Errorf("更新redis进度失败: %v", err)
		}
	}

	return nil
}

// buildPddGoodsSpecId 根据名称获取规格信息
// @param pddDll pddDLL对象
// @param token 授权令牌
// @param specId 商品规格id
// @param specName 规格名称
// @return DllGoodsSpec 规格信息
// @return error 错误信息
func buildPddGoodsSpecId(pddDll *pdd.PddDLL, token string, id string, name string) (_type.DllGoodsSpec, error) {

	var spec _type.DllGoodsSpec
	client, err := config.GetPddClient()
	if err != nil {
		return spec, err
	}
	//发送请求 生成商家自定义的规格
	clientId := client.ClientId
	clientSecret := client.ClientSecret
	specStr, err := pddDll.PddGoodsSpecIdGet(clientId, clientSecret, token, id, name)
	if err != nil {
		return spec, err
	}

	// 解析JSON字符串
	err = json.Unmarshal([]byte(specStr), &spec)
	if err != nil {
		return spec, fmt.Errorf("解析拼多多 PddGoodsSpecIdGet 接口返回json失败: %v [拼多多数据：%v]", err, specStr)
	}
	return spec, nil
}
