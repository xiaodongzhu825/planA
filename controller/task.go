package controller

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"planA/initialization/golabl"
	"planA/rep"
	serviceMysql "planA/service/mysql"
	"planA/tool/process"
	"planA/type/mysql"
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
	toolPdd "planA/tool/pdd"
	_type "planA/type"
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
	//将 imgTypeStr 转为 int64
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
	//将 updateTypeStr 转为 int64
	updateType := int64(0)
	if dataVal.UpdateType != "" {
		updateType, err = strconv.ParseInt(dataVal.UpdateType, 10, 64)
		if err != nil {
			errMsg := "更新方式转换失败: " + err.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	}
	// 查询店铺数据
	shopDataStr, err := service.GetTaskShop(dataVal.ShopID)
	if err != nil {
		errMsg := "获取店铺数据失败: shopId " + dataVal.ShopID + " " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 解析 json数据
	shopData, err := toolPdd.ParseShopData(shopDataStr)
	if err != nil {
		errMsg := "解析店铺数据失败:" + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 定义变量
	var spec *_type.Spec
	var context *_type.ShopContext
	var priceRange []_type.PriceRange
	// 实例化
	spec = &_type.Spec{}              // 指向零值结构体的指针
	context = &_type.ShopContext{}    // 指向零值结构体的指针
	priceRange = []_type.PriceRange{} // 空切片

	if shopData.Shop == nil {
		errMsg := "店铺数据为空"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	shop := shopData.Shop

	// 校验店铺订阅时间是否到期
	checkShopSubscriptionExpirationErr := checkShopSubscriptionExpiration(shop.ID, dataVal.ShopType, shop.SkuSpec, shop.Deregulation, shop.ExpirationTime)
	if checkShopSubscriptionExpirationErr != nil {
		errMsg := checkShopSubscriptionExpirationErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	if shopData.ShopDetail == nil {
		errMsg := "未设置商品详情"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	detail := shopData.ShopDetail

	if shop.ShopType != dataVal.ShopType {
		errMsg := "店铺类型不匹配 错误店铺类型:" + shop.ShopType
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//dataVal.ShopType = dataVal.ShopType

	//扣费
	//userId := strconv.FormatInt(shop.CreateBy, 10)
	//_, taskDeductionErr := TaskDeduction(shopID, userId)
	//if taskDeductionErr != nil {
	//	errMsg := "请求创建任务接口失败: " + taskDeductionErr.Error() + "店铺id：" + shopID + "用户id：" + userId
	//	tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
	//	return
	//}

	//验证店铺规格信息是否正确
	if dataVal.ShopType == "1" {
		pddDll, initPddSOErr := pdd.InitPddDll()
		if initPddSOErr != nil {
			errMsg := "初始化pdd.so失败: " + initPddSOErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//使用类目预测接口测试Token 是否有效
		err := toolPdd.BuildPddGoodsOuterCatMappingGet(pddDll, shop.Token)

		if err != nil && err.Error() == "拼多多Token已过期" {
			//更新token
			reqData := map[string]string{
				"shopId": dataVal.ShopID,
			}
			_, submitFormDataErr := tool.SubmitFormData(golabl.Config.FileUrl.UpdateTokenUrl, reqData)
			if submitFormDataErr != nil {
				fmt.Println("提交表单数据失败：", submitFormDataErr)
				return
			}
		} else if err != nil {
			tool.Error(httpMsg, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//请求创建任务接口并获取任务 id
	taskId, err := CreateTaskRequest(dataVal.ShopID, dataVal.TaskType)
	if err != nil {
		errMsg := "店铺ID " + dataVal.ShopID + " 任务类型" + dataVal.ShopType + " 请求创建任务接口失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//如果是拉取任务则校验是否已有在执行的  taskType == 3(拉取任务) || (taskType == 4 && dataVal.ShopType == "1")(拼多多拉取详情任务)
	if taskType == 3 || (taskType == 4 && dataVal.ShopType == "1") {
		//查询店铺拉取商品或者拉取商品详情所有任务
		read := rep.CreateDbFactoryRead()
		taskList, getTaskByShopIdAndTaskTypeErr := read.GetTaskByShopIdAndTaskType(shopData.Shop.ID, taskType)
		if getTaskByShopIdAndTaskTypeErr != nil {
			errMsg := "获取任务失败: " + getTaskByShopIdAndTaskTypeErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		for _, task := range taskList {
			// 获取 header信息
			taskHeader, getTaskHeaderErr := service.GetTaskHeader(task.TaskId)
			if getTaskHeaderErr != nil {
				errMsg := "获取任务头失败: " + getTaskHeaderErr.Error()
				tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
				return
			}
			//判断这个店铺是否已经存在正在执行的任务
			if taskHeader.Status == 1 || taskHeader.Status == 10 {
				errMsg := "当前店铺已经有此类任务正在执行中"
				tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
				return
			}
		}
	} else if taskType == 1 || taskType == 2 || taskType == 6 || taskType == 7 || taskType == 8 || taskType == 9 {
		if shopData.Spec == nil {
			errMsg := "未设置规格"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		spec = shopData.Spec
		//如果价格模版为空，则返回错误信息
		if shopData.PriceTemplate == nil {
			errMsg := "未选择价格模版"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		context = shopData.ShopContext
		//如果价格模版为空，则返回错误信息
		if shopData.PriceTemplate == nil {
			errMsg := "未选择价格模版"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		priceTemplateRangeStr := shopData.PriceTemplate.RangePrice
		err = json.Unmarshal([]byte(priceTemplateRangeStr), &priceRange)
		if err != nil {
			errMsg := "解析价格模板失败:" + err.Error() + " 原始数据：" + shopData.PriceTemplate.RangePrice
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}

		////验证店铺规格信息是否正确
		//if dataVal.ShopType == "1" && dataVal.TaskType == "1" {
		//	pddDll, initPddSOErr := pdd.InitPddDll()
		//	if initPddSOErr != nil {
		//		errMsg := "初始化pdd.so失败: " + initPddSOErr.Error()
		//		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		//		return
		//	}
		//	_, buildPddGoodsSpecIdErr := buildPddGoodsSpecId(pddDll, shop.Token, spec.SpecTypeID, spec.SpecName)
		//	if buildPddGoodsSpecIdErr != nil {
		//		errMsg := "构建规格ID失败: " + buildPddGoodsSpecIdErr.Error()
		//		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		//		return
		//	}
		//}
	}

	fmt.Printf("店铺ID: %s, 店铺类型: %s, 任务类型: %s, 更新方式: %s, 任务数量: %s 任务id: %s \n", dataVal.ShopID, dataVal.ShopType, dataVal.TaskType, dataVal.UpdateType, dataVal.TaskCount, taskId)

	// 创建任务逻辑...
	createAt := time.Now().Unix()
	task, err := CreateTaskData(taskId, taskType, createAt, shop, priceRange, spec, detail, context, dataVal.TaskCount, imgType, updateType, shopData.PriceTemplate.PriceType)
	if err != nil {
		errMsg := "创建任务失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//推送 redis
	err = service.UpdateTaskHeader(taskId, task.Header)
	if err != nil {
		errMsg := "保存任务头失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//更新任务尾
	err = service.UpdateTaskFooter(taskId, &task.Footer)
	if err != nil {
		errMsg := "保存任务尾失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	if taskType == 10 || taskType == 11 {
		var createDelTask mysql.DelTask
		userId := shop.CreateBy
		// 将 task.Header 转为 json字符串
		headerJson, marshalErr := json.Marshal(task.Header)
		if marshalErr != nil {
			errMsg := "任务头转换json失败: " + marshalErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		headerStr := string(headerJson)
		taskCountOver := 0
		status := 0

		if taskType == 10 {
			taskType := 2
			//将 DelNum 转为 int64
			taskCount, err := strconv.ParseInt(dataVal.DelNum, 10, 64)
			if err != nil {
				errMsg := "任务类型转换失败: " + err.Error()
				tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
				return
			}
			taskCountInt := int(taskCount)
			createDelTask = mysql.DelTask{
				UserID:        &userId,
				ShopID:        &shop.ID,
				TaskID:        &taskId,
				TaskType:      &taskType,
				ShopType:      &shop.ShopType,
				Status:        &status,
				ShopName:      &shop.ShopName,
				TaskCountOver: &taskCountOver,
				TaskCount:     &taskCountInt,
				Header:        &headerStr,
				CreateAt:      nil,
			}
		} else {
			taskType := 3
			taskCount := 0
			// 将时间字符串转换为时间戳 int64
			layout := "2006-01-02 15:04:05"
			loc, _ := time.LoadLocation("Asia/Shanghai") // 北京时间
			t, err := time.ParseInLocation(layout, dataVal.DelTime, loc)
			if err != nil {
				errMsg := "时间解析失败: " + err.Error()
				tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
				return
			}
			delTimeInt64 := t.Unix()

			// 转换为 time.Time 类型
			delTimeObj := time.Unix(delTimeInt64, 0)
			createDelTask = mysql.DelTask{
				UserID:        &userId,
				ShopID:        &shop.ID,
				TaskID:        &taskId,
				TaskType:      &taskType,
				ShopType:      &shop.ShopType,
				Status:        &status,
				ShopName:      &shop.ShopName,
				TaskCountOver: &taskCountOver,
				TaskCount:     &taskCount,
				Header:        &headerStr,
				StopAt:        &delTimeObj,
				CreateAt:      nil,
			}
		}
		err = serviceMysql.CreateDelTask(createDelTask)
		if err != nil {
			errMsg := "创建删除任务失败: " + err.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}

	} else if taskType == 3 || (taskType == 4 && dataVal.ShopType == "1") {
		//如果是拉取任务则直接执行 B方法程序  taskType == 3(拉取任务) || (taskType == 4 && dataVal.ShopType == "1")(拼多多拉取详情任务)
		// 执行 B方法程序
		_, runTaskWorkerErr := process.RunTaskWorker(taskId)
		if runTaskWorkerErr != nil {
			fmt.Printf("执行B程序出错: %v\n", runTaskWorkerErr)
			return
		}
		mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()

		//写入 mysql数据
		mysqlCreateTaskRecordsErr := mysqlWrite.CreateTaskRecords(_type.TaskRecordsDTO{
			UserId:   shopData.Shop.CreateBy,
			ShopId:   shopData.Shop.ID,
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
			UserId:   shopData.Shop.CreateBy,
			ShopId:   shopData.Shop.ID,
			TaskId:   taskId,
			ShopName: shop.ShopName,
			TaskType: taskType,
		})
		if sqliteTaskExportErr != nil {
			errMsg := "插入任务用户失败: " + sqliteTaskExportErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	} else {
		mysqlWrite, sqliteWrite := rep.CreateDbFactoryWrite()

		//写入 mysql数据
		mysqlCreateTaskRecordsErr := mysqlWrite.CreateTaskRecords(_type.TaskRecordsDTO{
			UserId:   shopData.Shop.CreateBy,
			ShopId:   shopData.Shop.ID,
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
			UserId:   shopData.Shop.CreateBy,
			ShopId:   shopData.Shop.ID,
			TaskId:   taskId,
			ShopName: shop.ShopName,
			TaskType: taskType,
		})
		if sqliteTaskExportErr != nil {
			errMsg := "插入任务用户失败: " + sqliteTaskExportErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	}

	// 返回成功响应
	tool.Success(httpMsg, taskId)
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
	if header.TaskId == "" {
		errMsg := "任务不存在或已经删除"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 更新任务数
	go UpdateTaskCount(bodyData, taskId)

	// 返回成功响应
	tool.Success(httpMsg, "")
}

// PauseTask 暂停任务
func PauseTask(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.TaskIdValidator(data)
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
	tool.Success(httpMsg, "")
}

// ResumeTask 恢复任务
func ResumeTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.TaskIdValidator(data)
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

	// 恢复 B程序
	suspendProcessErr := process.ResumeProcess(dataVal.TaskID)
	if suspendProcessErr != nil {
		errMsg := "恢复进程失败: " + suspendProcessErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 返回成功响应
	tool.Success(httpMsg, "")
}

// StopTask 停止任务
func StopTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.TaskIdValidator(data)
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
	tool.Success(httpMsg, "")
}

// DelTask 删除任务
func DelTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.TaskIdValidator(data)
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
		errMsg := "删除任务失败: " + mysqlDeleteTaskRecordsByTaskIdErr.Error()
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

	tool.Success(httpMsg, "")
}

// OverTask 任务完成
func OverTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, updateTaskStatusValidatorErr := validator.TaskIdValidator(data)
	if updateTaskStatusValidatorErr != nil {
		tool.Error(httpMsg, updateTaskStatusValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	//查询 header 信息
	header, getTaskHeaderErr := service.GetTaskHeader(dataVal.TaskID)
	if getTaskHeaderErr != nil {
		fmt.Printf("获取footer 信息失败 %v", getTaskHeaderErr)
		return
	}
	if header.Status != _type.TaskStatusStopped {
		//推送 redis
		status := int64(_type.TaskStatusOver)
		err := service.UpdateHeaderStatus(dataVal.TaskID, status)
		if err != nil {
			errMsg := err.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		lock.DestroyLock(dataVal.TaskID) //销毁锁
	}
	if header.TaskType == 5 {
		taskNoticeRequestErr := OperationGoodsTaskNoticeRequest(header.TaskId, header.ShopId)
		if taskNoticeRequestErr != nil {
			return
		}
	} else {
		taskNoticeRequestErr := TaskNoticeRequest(header.TaskId)
		if taskNoticeRequestErr != nil {
			return
		}
	}
	// 返回成功响应
	tool.Success(httpMsg, "")
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
		bodyOver, _, GetTaskBodyOverErr := service.GetTaskBodyOver(v.TaskId, 0, 10)
		if GetTaskBodyOverErr != nil {
			errMsg := fmt.Sprintf("获取body_over 信息失败 %v", GetTaskBodyOverErr)
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
		header.ShopMsg.Token = "****暂不展示*****"
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
	tool.Success(httpMsg, dataRet)
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
		bodyOver, _, GetTaskBodyOverErr := service.GetTaskBodyOver(v.TaskId, 0, 10)
		if GetTaskBodyOverErr != nil {
			errMsg := fmt.Sprintf("获取body_over 信息失败 %v", GetTaskBodyOverErr)
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
	tool.Success(httpMsg, dataRet)
}

// GetTaskHeader 获取 header信息
func GetTaskHeader(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, getHeaderValidatorErr := validator.TaskIdValidator(data)
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
	//判断数据是否为空
	if header.TaskId == "" {
		tool.Success(httpMsg, "")
		return
	}
	header.ShopMsg.Token = "****暂不展示*****"
	tool.Success(httpMsg, header)
}

// GetBodyOver 获取body_over
func GetBodyOver(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, getBodyOverValidatorValidatorErr := validator.GetBodyOverValidator(data)
	if getBodyOverValidatorValidatorErr != nil {
		tool.Error(httpMsg, getBodyOverValidatorValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	// 获取分页参数
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	bodyOver, total, getTaskBodyOverLimit10Err := service.GetTaskBodyOver(dataVal.TaskID, page, size)
	if getTaskBodyOverLimit10Err != nil {
		errMsg := getTaskBodyOverLimit10Err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  bodyOver,
	}
	tool.Success(httpMsg, dataRet)
}

// GetTaskList 获取任务列表
func GetTaskList(httpMsg http.ResponseWriter, data *http.Request) {

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
	tool.Success(httpMsg, dataRet)
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
	tool.Success(httpMsg, "")
}

//****************************工具**************************************//

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
func CreateTaskData(taskId string, taskType int64, createAt int64, shop *_type.Shop, priceRange []_type.PriceRange, spec *_type.Spec, detail *_type.ShopDetail, context *_type.ShopContext, taskCount string, imgType int64, updateType int64, priceType string) (*_type.Task, error) {
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
	var token string
	var districtId int64
	var districtType string
	//处理 Token
	if shop.ShopType == "1" || shop.ShopType == "2" { //拼多店铺
		token = shop.Token
	} else if shop.ShopType == "5" { // 闲鱼店铺
		token = fmt.Sprintf("{\"app_id\":%v,\"app_secret\":\"%v\",\"username\":\"%v\"}", shop.MallID, shop.Token, shop.ShopKey)
		districtId = detail.DistrictId
		districtType = detail.DistrictType
	}
	// specTypeID 转换为int64
	var specTypeID int64
	var parseAdjustPercentErr error
	if spec.SpecTypeID != "" {
		specTypeID, parseAdjustPercentErr = parseAdjustPercent(spec.SpecTypeID)
		if parseAdjustPercentErr != nil {
			return &task, fmt.Errorf("规格类型ID 转换失败: %v", parseAdjustPercentErr)
		}
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
				ID:                          detail.ID,                                              //店铺详情 ID
				ShopAliasName:               shop.ShopName,                                          //店铺别名
				ShopName:                    shop.ShopName,                                          //店铺名称
				Token:                       token,                                                  //店铺 token【如果是咸鱼店铺，此token则是应用密钥】
				GoodsNamePrefix:             detail.TitlePrefix,                                     //商品名称前缀
				GoodsNameSuffix:             detail.TitleSuffix,                                     //商品名称后缀
				TitleConsistOf:              detail.TitleConsistOf,                                  //商品名称组成
				SpaceCharacter:              detail.SpaceCharacter,                                  //间隔字符  0无间隔 1空格
				WatermarkImgUrl:             detail.WatermarkImgUrl,                                 //水印图片
				WatermarkPosition:           detail.WatermarkPosition,                               //水印位置 0全部  1第一张
				CarouseLastImgUrlArray:      tool.FilterStrings(detail.CarouseLastImgUrlArray),      //轮播图最后图片[]string（tool.FilterStrings 函数为去掉数组中的空、图片不合法等字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				GoodsDetailFirstImgUrlArray: tool.FilterStrings(detail.GoodsDetailFirstImgUrlArray), //商品详情首图URL数组[]string（tool.FilterStrings 函数为去掉数组中的空、图片不合法字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				GoodsDetailLastImgUrlArray:  tool.FilterStrings(detail.GoodsDetailLastImgUrlArray),  //商品详情最后图片URL数组（tool.FilterStrings 函数为去掉数组中的空、图片不合法等字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				IsFolt:                      detail.Fake == "1",                                     //是否支持假一赔十，false-不支持，true-支持
				IsPreSale:                   detail.Presale == "1" || detail.Presale == "2",         //是否预售,true-预售商品，false-非预售商品
				IsRefundable:                detail.SevenDays == "1",                                //是否7天无理由退换货，true-支持，false-不支持
				ShipmentLimitSecond:         shipmentLimitSecond,                                    //承诺发货时间（秒）
				CostTemplateId:              detail.TemplateId,                                      //物流运费模板 ID
				SpecName:                    spec.SpecTypeName,                                      //规格名称
				SpecId:                      specTypeID,                                             //规格 ID
				SpecChildName:               spec.SpecName,                                          //规格子名称
				DefStock:                    int64(detail.StockDeff),                                //默认库存
				TwoDiscount:                 detail.TowDiscount,                                     //2折
				IsSecondHand:                detail.IsSecondHand == "1",                             //是否二手 1 -二手商品 ，0-全新商品
				DistrictMsg: _type.DistrictMsg{
					DistrictId:   districtId,
					DistrictType: districtType,
				},
				ShopContext:        context.Context,           //店铺描述
				SkuWatermarkImgUrl: detail.SkuWatermarkImgUrl, //sku 水印图片
				PublishType:        detail.PublishType,        //发布方式 0=24（图书类目） 1=99（其他类目）【限闲鱼店铺使用】
				CategoryId:         detail.CategoryId,         //类目 Id【限闲鱼店铺使用】
				SpecCompose:        spec.SpecCompose,          //规格组合类型 0=自定义 1=Isbn 2=书名 3=货号
				SpecPrefix:         spec.SpecPrefix,           //规格前缀
				SpecSuffix:         spec.SpecSuffix,           //规格后缀
				IsParcel:           detail.IsParcel,           //是否包邮
				BookWeight:         detail.BookWeight,         //书籍重量
				StandardNumber:     detail.StandardNumber,     //商品标准本数
				ConditionDef:       detail.ConditionDef,       //商品品相
			},
			PriceMod:         priceModArr,             //价格模版
			PriceType:        priceType,               //价格类型
			ShipPriceMod:     "",                      //运费模版
			TaskCount:        taskCountInt64,          //任务数量
			TaskCountTrue:    0,                       //真实任务数量
			TaskCountWait:    0,                       //等待任务数量
			TaskCountOver:    0,                       //任务完成数量
			TaskCountSuccess: 0,                       //任务成功数量
			TaskCountError:   0,                       //任务失败数量
			Status:           _type.TaskStatusRunning, //任务状态 1=运行中 2=暂停中 3=完成
			TaskQpm:          0,                       //任务 QPM
			TaskCreateAt:     createAt,                //任务创建时间
			TaskOverAt:       0,                       //任务完成时间
			LastIndex:        0,                       //最后索引
			ImgType:          imgType,                 //图片类型 0=无图片 1=轮播图 2=商品详情首图 3=商品详情最后图片
			UpdateType:       updateType,              // 更新方式（仅核价发布或核价表格发布使用） 1 过滤重复 2 全新上传
			Pool: _type.PoolConfig{ //协程池配置
				Size:                 500,  //协程数量
				WithExpiryDuration:   10,   //过期时间
				WithPreAlloc:         true, //预分配
				WithMaxBlockingTasks: 2000, //阻塞任务数
				WithNonblocking:      true, //非阻塞
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
		fmt.Printf("找到的书品为0，所以不提交到redis %v", bodyData)
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
			fmt.Printf("解析失败: %v %v\n", err, jsonStr)
			continue
		}
		var bookInfo _type.BookInfo
		var GetTaskBookErr error
		// 书品处理
		if header.TaskType == 1 || header.TaskType == 2 || header.TaskType == 5 || header.TaskType == 6 || header.TaskType == 7 || header.TaskType == 8 || header.TaskType == 9 {
			// 连接DB[b] 获取书品信息,#操作商品的isbn13个0则不查询isbn
			if !(header.TaskType == 5 && taskBody.BookInfo.Isbn == "0000000000000") {
				bookInfo, GetTaskBookErr = service.GetTaskBook(taskBody.BookInfo.Isbn)
				if GetTaskBookErr != nil {
					if errors.Is(GetTaskBookErr, _redis.Nil) {
						setNoBookCountErr := service.SetNoBookCount(taskBody.BookInfo.Isbn)
						if setNoBookCountErr != nil {
							fmt.Printf("设置无书品数量失败 isbn:%v", taskBody.BookInfo.Isbn)
						}
					}
					if header.TaskType != 5 {
						fmt.Printf("获取BookInfo失败-原因: %v\n", GetTaskBookErr)
						continue
					}
				}
			}
			// 图片处理
			if header.TaskType == 1 || header.TaskType == 2 || header.TaskType == 6 || header.TaskType == 7 || header.TaskType == 8 || header.TaskType == 9 {
				//处理图片 仅官图不处理
				if header.ImgType == 2 { //仅实拍图，使用传递过来的图片
					bookInfo.ImageObject.CarouselUrlArray = taskBody.BookInfo.ImageObject.CarouselUrlArray
				} else if header.ImgType == 3 { // 优先官图，优先使用 bookInfo中的图片，如果没有使用传递过来的图片
					if len(bookInfo.ImageObject.CarouselUrlArray) == 0 {
						bookInfo.ImageObject.CarouselUrlArray = taskBody.BookInfo.ImageObject.CarouselUrlArray
					}
				} else if header.ImgType == 4 { //优先实拍，优先使用 传递过来的图片，如果没有使用bookInfo中的图片
					if len(taskBody.BookInfo.ImageObject.CarouselUrlArray) > 0 {
						bookInfo.ImageObject.CarouselUrlArray = taskBody.BookInfo.ImageObject.CarouselUrlArray
					}
				}

				// 类目 Id处理
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
			}
			//表格上传处理
			if header.TaskType == 2 {
				// 书名处理，如果表格上传存在，则使用表格中的
				if taskBody.BookInfo.BookName != "" {
					bookInfo.BookName = taskBody.BookInfo.BookName
				}
				// 图片处理，如果表格上传存在，则使用表格中的
				if len(taskBody.BookInfo.ImageObject.CarouselUrlArray) > 0 {
					bookInfo.ImageObject.CarouselUrlArray = taskBody.BookInfo.ImageObject.CarouselUrlArray
				}
			}
			// 更新 BookInfo
			taskBody.BookInfo = bookInfo
		}

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
	if header.TaskType == 5 {
		taskNoticeRequestErr := OperationGoodsTaskNoticeRequest(taskId, header.ShopId)
		if taskNoticeRequestErr != nil {
			return 0
		}
	} else {
		taskNoticeRequestErr := TaskNoticeRequest(taskId)
		if taskNoticeRequestErr != nil {
			return 0
		}
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
func CreateTaskRequest(shopId string, taskType string) (string, error) {

	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		errMsg := "获取文件路径配置失败: " + getFileUrlConfigErr.Error()
		return "", fmt.Errorf(errMsg)
	}
	taskTypeName := tool.GetTaskTypeName(taskType)
	dataMap := map[string]string{
		"shopId":   shopId,
		"taskType": "NEW_ADD_TASK",
		"fileName": taskTypeName,
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

// OperationGoodsTaskNoticeRequest 操作商品任务有等待数据通知接口
func OperationGoodsTaskNoticeRequest(taskId string, shopId string) error {

	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		return fmt.Errorf("获取文件路径配置失败: %v", getFileUrlConfigErr)
	}
	data := map[string]string{
		"taskId": taskId,
		"shopId": shopId,
	}
	_, submitFormDataErr := tool.SubmitFormData(fileUrlConfig.CreateOperationTaskNoticeUrl, data)
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
// @param taskType 任务类型
// @return error
func AppendToCSV(fileName string, data []_type.TaskBody, writeHeader bool, taskId string, taskType int64) error {

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
		errStr := item.Detail.Error
		if taskType == 3 || taskType == 4 {
			errStr = ""
		}
		record := []string{
			item.BookInfo.Isbn,
			item.BookInfo.BookName,
			statusStr,
			errStr,
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

// 验证店铺订阅是否到期
func checkShopSubscriptionExpiration(taskId string, shopType string, skuSpec string, deregulation string, expirationTime string) error {
	if shopType == "2" || shopType == "5" {
		// 检验店铺订阅时间是否到期
		expirationTime, err := tool.GetSubscriptionExpirationTime(taskId)
		if err != nil {
			return fmt.Errorf("获取用户订阅到期时间失败: %v", err)
		}
		// 明确单位：假设 expirationTime 是秒级时间戳
		now := time.Now().Unix()
		if now > expirationTime {
			expirationDateTime := time.Unix(expirationTime, 0).Format("2006-01-02 15:04:05")
			return fmt.Errorf("店铺已到期，到期时间：" + expirationDateTime)
		}
		return nil
	} else {
		// 解析时间字符串
		expTime, err := time.Parse("2006-01-02 15:04:05", expirationTime)
		if err != nil {
			return fmt.Errorf("时间格式错误: %v", err)
		}

		// 判断是否大于当前时间
		if !expTime.After(time.Now()) {
			return fmt.Errorf("店铺已到期，到期时间：%s", expTime.Format("2006-01-02 15:04:05"))
		}
		// 如果是拼多多店铺的话校验下 是否试用
		if shopType == "1" && strings.Contains(skuSpec, "7天") {
			// 是否开启使用无上限
			if deregulation == "1" {
				return fmt.Errorf("无法创建任务，请去ERP店铺列表订阅或者开通七天无上限限免")
			}
		}
		return nil
	}
}
