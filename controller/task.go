package controller

import (
	"bytes"
	"encoding/csv"
	"encoding/json"

	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"planA/controlState/lock"
	"planA/initialization/config"
	"planA/modules/logs"
	"planA/modules/pdd"
	"planA/service"
	"planA/tool"
	"planA/type"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	mysqlType "planA/type/mysql"
	redisType "planA/type/redis"
	sqLiteType "planA/type/sqLite"
)

var expiration = 8 * 24 * time.Hour
var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	modntdll    = syscall.NewLazyDLL("ntdll.dll")

	procOpenProcess      = modkernel32.NewProc("OpenProcess")
	procCloseHandle      = modkernel32.NewProc("CloseHandle")
	procNtSuspendProcess = modntdll.NewProc("NtSuspendProcess")
	procNtResumeProcess  = modntdll.NewProc("NtResumeProcess")
)

const (
	PROCESS_SUSPEND_RESUME = 0x0800
)

// CreateTask 创建任务
func CreateTask(httpMsg http.ResponseWriter, data *http.Request) {
	// 获取表单数据
	shopID := data.FormValue("shop_id")
	shopType := data.FormValue("shop_type")
	taskCount := data.FormValue("task_count")
	imgTypeStr := data.FormValue("img_type")
	imgType, err := strconv.ParseInt(imgTypeStr, 10, 64)
	if err != nil {
		errMsg := "图片类型转换失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	taskTypeStr := data.FormValue("task_type")
	if taskTypeStr == "" {
		taskTypeStr = "1"
	}
	//将 taskTypeStr 转为 int64
	taskType, err := strconv.ParseInt(taskTypeStr, 10, 64)
	if err != nil {
		errMsg := "任务类型转换失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 1. 验证店铺 ID
	if shopID == "" {
		errMsg := "店铺ID不能为空"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 验证店铺类型
	if shopType != "1" && shopType != "2" && shopType != "5" {
		errMsg := "店铺类型错误"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 验证任务数
	if taskCount == "" {
		errMsg := "任务数不能为空"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//验证任务类型
	if taskType != 1 && taskType != 2 {
		errMsg := "任务类型错误"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//验证图片类型
	if imgType != 1 && imgType != 2 && imgType != 3 && imgType != 4 {
		errMsg := "图片类型错误"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 查询店铺数据
	shopDataStr, err := service.GetTaskShop(shopID)
	if err != nil {
		errMsg := "获取店铺数据失败: shopId " + shopID + " " + err.Error()
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
	priceTemplateRangeStr := shopData.PriceTemplate.RangePrice

	var priceRange []_type.PriceRange
	err = json.Unmarshal([]byte(priceTemplateRangeStr), &priceRange)
	if err != nil {
		errMsg := "解析价格模板失败:" + err.Error() + " 原始数据：" + shopData.PriceTemplate.RangePrice
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	if shop.ShopType != shopType {
		errMsg := "店铺类型不匹配 错误店铺类型:" + shop.ShopType
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//验证店铺规格信息是否正确
	//pddDll, initPddSOErr := pdd.InitPddSO()
	//if initPddSOErr != nil {
	//	errMsg := "初始化pdd.so失败: " + initPddSOErr.Error()
	//	tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
	//	return
	//}
	//_, buildPddGoodsSpecIdErr := buildPddGoodsSpecId(pddDll, shop.Token, spec.SpecTypeID, spec.SpecName)
	//if buildPddGoodsSpecIdErr != nil {
	//	errMsg := "构建规格ID失败: " + buildPddGoodsSpecIdErr.Error()
	//	tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
	//	return
	//}

	//扣费
	//userId := strconv.FormatInt(shop.CreateBy, 10)
	//_, taskDeductionErr := TaskDeduction(shopID, userId)
	//if taskDeductionErr != nil {
	//	errMsg := "请求创建任务接口失败: " + taskDeductionErr.Error() + "店铺id：" + shopID + "用户id：" + userId
	//	tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
	//	return
	//}

	//请求创建任务接口并获取任务 id
	taskId, err := CreateTaskRequest(shopID)
	if err != nil {
		errMsg := "请求创建任务接口失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	fmt.Printf("店铺ID: %s, 店铺类型: %s, 任务数量: %s 任务id: %s \n", shopID, shopType, taskCount, taskId)

	// 创建任务逻辑...
	createAt := time.Now().Unix()
	task, err := CreateTaskData(taskId, taskType, createAt, shop, priceRange, spec, detail, taskCount, imgType)
	if err != nil {
		errMsg := "创建任务失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//推送redis
	err = service.UpdateTaskHeader(taskId, task.Header, expiration)
	if err != nil {
		errMsg := "保存任务头失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	err = service.UpdateTaskFooter(taskId, &task.Footer, expiration)
	if err != nil {
		errMsg := "保存任务尾失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//写入 sqlite
	insertTaskRecordErr := service.InsertTaskRecord(sqLiteType.TaskRecord{
		UserID:   shopData.Shop.CreateBy,
		TaskID:   taskId,
		ShopName: shop.ShopName,
		TaskType: taskType,
	})
	if insertTaskRecordErr != nil {
		errMsg := "插入任务记录失败: " + insertTaskRecordErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//写入 mysql
	insertTaskUserErr := service.InsertTaskUser(&mysqlType.TaskUser{
		UserID:   &shopData.Shop.CreateBy,
		ShopID:   &shopData.Shop.ID,
		TaskID:   &taskId,
		ShopName: &shop.ShopName,
		TaskType: &taskType,
	})
	if insertTaskUserErr != nil {
		errMsg := "插入任务用户失败: " + insertTaskUserErr.Error()
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

	// 从路径参数获取 id
	vars := mux.Vars(data)
	taskId := vars["id"]
	// 1. 验证任务 ID
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
	if header.Status != _type.TaskStatusRunning {
		errMsg := "当前状态不是执行中"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//推送 redis
	status := int64(_type.TaskStatusPaused)
	err := service.UpdateHeaderStatus(taskId, status, expiration)
	if err != nil {
		errMsg := err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 暂停 B程序
	headerKey := fmt.Sprintf("%s:header", taskId)
	processId, getProcessIdErr := service.GetProcessId(headerKey)
	if getProcessIdErr != nil {
		errMsg := "获取进程号失败: " + getProcessIdErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	suspendProcessErr := SuspendProcess(processId)
	if suspendProcessErr != nil {
		errMsg := "暂停进程失败: " + suspendProcessErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 返回成功响应
	tool.Session(httpMsg, "")
}

// ResumeTask 恢复任务
func ResumeTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 从路径参数获取 id
	vars := mux.Vars(data)
	taskId := vars["id"]
	// 1. 验证任务 ID
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
	if header.Status != _type.TaskStatusPaused {
		errMsg := "当前状态不是暂停"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//推送redis
	status := int64(_type.TaskStatusRunning)
	err := service.UpdateHeaderStatus(taskId, status, expiration)
	if err != nil {
		errMsg := err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 恢复 B程序
	headerKey := fmt.Sprintf("%s:header", taskId)
	processId, getProcessIdErr := service.GetProcessId(headerKey)
	if getProcessIdErr != nil {
		errMsg := "获取进程号失败: " + getProcessIdErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	suspendProcessErr := ResumeProcess(processId)
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

	// 从路径参数获取 id
	vars := mux.Vars(data)
	taskId := vars["id"]
	// 1. 验证任务 ID
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
	if header.Status == _type.TaskStatusOver {
		errMsg := "任务已完成"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 恢复 B程序
	headerKey := fmt.Sprintf("%s:header", taskId)
	processId, getProcessIdErr := service.GetProcessId(headerKey)
	if getProcessIdErr != nil {
		errMsg := "获取进程号失败: " + getProcessIdErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	suspendProcessErr := ResumeProcess(processId)
	if suspendProcessErr != nil {
		errMsg := "恢复进程失败: " + suspendProcessErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//修改 Header中的状态 并且 删除 bodyWait 中的数据
	err := service.StopTask(taskId, expiration)
	if err != nil {
		errMsg := err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 返回成功响应
	tool.Session(httpMsg, "")
}

// OverTask 任务完成
func OverTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 从路径参数获取 id
	vars := mux.Vars(data)
	taskId := vars["id"]
	// 1. 验证任务 ID
	if taskId == "" {
		errMsg := "任务 ID 不能为空"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//推送 redis
	status := int64(_type.TaskStatusOver)
	err := service.UpdateHeaderStatus(taskId, status, expiration)
	if err != nil {
		errMsg := err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	lock.DestroyLock(taskId) //销毁锁
	// 返回成功响应
	tool.Session(httpMsg, "")
}

// GetTask 任务列表
func GetTask(httpMsg http.ResponseWriter, data *http.Request) {

	// 获取分页参数
	page, size := tool.SetPage(data.URL.Query().Get("page"), data.URL.Query().Get("size"))
	taskId := data.URL.Query().Get("task_id")
	shopName := data.URL.Query().Get("shop_name")
	taskTypeStr := data.URL.Query().Get("task_type")
	taskTypeInt := 0
	var atoiErr error
	if taskTypeStr != "" {
		//将 taskTypeStr 转为 int
		taskTypeInt, atoiErr = strconv.Atoi(taskTypeStr)
		if atoiErr != nil {
			errMsg := "任务类型转换失败"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}

	}

	records, total, err := service.GetTaskRecordsWithPage(page, size, taskId, shopName, taskTypeInt)
	if err != nil {
		errMsg := err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	dataTaskAll := []map[string]interface{}{}
	for _, v := range records {
		//查询 header 信息
		header, getTaskHeaderErr := service.GetTaskHeader(v.TaskID)
		if getTaskHeaderErr != nil {
			errMsg := fmt.Sprintf("获取footer 信息失败 %v", getTaskHeaderErr)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//获取 footer 信息
		footer, getTaskFooterErr := service.GetTaskFooter(v.TaskID)
		if getTaskFooterErr != nil {
			errMsg := fmt.Sprintf("获取footer 信息失败 %v", getTaskFooterErr)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//获取 body_over 信息
		bodyOver, GetTaskBodyOverLimit10Err := service.GetTaskBodyOverLimit10(v.TaskID)
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

// GetTaskByUserId 获取用户任务
func GetTaskByUserId(httpMsg http.ResponseWriter, data *http.Request) {
	// 获取分页参数
	page, size := tool.SetPage(data.URL.Query().Get("page"), data.URL.Query().Get("size"))
	userIdStr := data.URL.Query().Get("user_id")
	if userIdStr == "" {
		errMsg := "用户ID 不能为空"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	userId, parseIntuserIdErr := strconv.ParseInt(userIdStr, 10, 64)
	if parseIntuserIdErr != nil {
		errMsg := "用户ID 转换失败"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	taskId := data.URL.Query().Get("task_id")
	shopName := data.URL.Query().Get("shop_name")
	taskType := data.URL.Query().Get("task_type")
	taskTypeInt64 := int64(0)
	var parseIntTaskTypeErr error
	if taskType != "" {
		//将taskType 转换为 int64
		taskTypeInt64, parseIntTaskTypeErr = strconv.ParseInt(taskType, 10, 64)
		if parseIntTaskTypeErr != nil {
			errMsg := "任务类型转换失败"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	}
	records, total, PageQueryTaskUserByUserIdErr := service.PageQueryTaskUserByUserId(&mysqlType.PageQueryTaskUserByUserIdParams{
		Page: _type.Page{
			PageNum:  page,
			PageSize: size,
		},
		ShopName: shopName,
		TaskID:   taskId,
		UserID:   userId,
		TaskType: taskTypeInt64,
	})
	if PageQueryTaskUserByUserIdErr != nil {
		errMsg := PageQueryTaskUserByUserIdErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	dataTaskAll := []map[string]interface{}{}
	for _, v := range records {
		//查询 header 信息
		header, getTaskHeaderErr := service.GetTaskHeader(*v.TaskID)
		if getTaskHeaderErr != nil {
			errMsg := fmt.Sprintf("获取footer 信息失败 %v", getTaskHeaderErr)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//获取 footer 信息
		footer, getTaskFooterErr := service.GetTaskFooter(*v.TaskID)
		if getTaskFooterErr != nil {
			errMsg := fmt.Sprintf("获取footer 信息失败 %v", getTaskFooterErr)
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		//获取 body_over 信息
		bodyOver, GetTaskBodyOverLimit10Err := service.GetTaskBodyOverLimit10(*v.TaskID)
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

func B(httpMsg http.ResponseWriter, data *http.Request) {
	taskID := "111"
	_, callSendPublishingErr := callSendPublishing(taskID, "")
	if callSendPublishingErr != nil {
		logStr := fmt.Sprintf("执行B程序失败: [taskId] %v [error] %v", taskID, callSendPublishingErr.Error())
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, logStr)
		return
	}
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
// @param taskCount 任务数量
// @param detail 店铺详情
// @param imgType 图片类型
// @return *_type.Task 任务数据
// @return error 错误
func CreateTaskData(taskId string, taskType int64, createAt int64, shop *_type.Shop, priceRange []_type.PriceRange, spec *_type.Spec, detail *_type.ShopDetail, taskCount string, imgType int64) (*_type.Task, error) {
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
	var goodsDetailFirstImgUrlArray []string
	var carouseLastImgUrlArray []string
	var goodsDetailLastImgUrlArray []string
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
				ID:                          detail.ID,                                       //店铺详情 ID
				ShopAliasName:               shop.ShopName,                                   //店铺别名
				ShopName:                    shop.ShopName,                                   //店铺名称
				Token:                       token,                                           //店铺 token【如果是咸鱼店铺，此token则是应用密钥】
				GoodsNamePrefix:             detail.TitlePrefix,                              //商品名称前缀
				GoodsNameSuffix:             detail.TitleSuffix,                              //商品名称后缀
				TitleConsistOf:              detail.TitleConsistOf,                           //商品名称组成
				SpaceCharacter:              detail.SpaceCharacter,                           //间隔字符  0无间隔 1空格
				WatermarkImgUrl:             detail.WatermarkImgUrl,                          //水印图片
				CarouseLastImgUrlArray:      tool.FilterStrings(carouseLastImgUrlArray),      //轮播图最后图片[]string（tool.FilterStrings 函数为去掉数组中的空、图片不合法等字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				GoodsDetailFirstImgUrlArray: tool.FilterStrings(goodsDetailFirstImgUrlArray), //商品详情首图URL数组[]string（tool.FilterStrings 函数为去掉数组中的空、图片不合法字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				GoodsDetailLastImgUrlArray:  tool.FilterStrings(goodsDetailLastImgUrlArray),  //商品详情最后图片URL数组（tool.FilterStrings 函数为去掉数组中的空、图片不合法等字符串，因为原始数据中可能会出现空字符串导致商品发布报图片信息错误）
				IsFolt:                      detail.Fake == "1",                              //是否支持假一赔十，false-不支持，true-支持
				IsPreSale:                   detail.Presale == "1" || detail.Presale == "2",  //是否预售,true-预售商品，false-非预售商品
				IsRefundable:                detail.SevenDays == "1",                         //是否7天无理由退换货，true-支持，false-不支持
				ShipmentLimitSecond:         shipmentLimitSecond,                             //承诺发货时间（秒）
				CostTemplateId:              int64(detail.TemplateId),                        //物流运费模板 ID
				SpecName:                    spec.SpecTypeName,                               //规格名称
				SpecId:                      specTypeID,                                      //规格 ID
				SpecChildName:               spec.SpecName,                                   //规格子名称
				DefStock:                    int64(detail.StockDeff),                         //默认库存
				TwoDiscount:                 detail.TowDiscount,                              //2折
				IsSecondHand:                detail.IsSecondHand == "1",                      //是否二手 1 -二手商品 ，0-全新商品
				DistrictMsg: _type.DistrictMsg{
					DistrictId:   districtId,
					DistrictType: districtType,
				},
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
	//如果没上锁，则启动A程序
	if lock.GetLock(taskId) {
		AddTask(taskId, bodyData)
	} else {
		//如果最后找到的书品数量为0，则不提交到redis
		if AddTask(taskId, bodyData) <= 0 {
			fmt.Println("找到的书品为0，所以不提交到redis")
			return
		}
		//如果没上锁，则启动B程序
		//if !lock.GetLock(taskId) {
		//	lock.SetLock(taskId) //设置锁
		//	// 执行 B方法程序
		//	headerKey := fmt.Sprintf("%s:header", taskId)
		//	processId, getProcessIdErr := service.GetProcessId(headerKey)
		//	//检查是否有错误
		//	if getProcessIdErr != nil && !errors.Is(getProcessIdErr, _redis.Nil) {
		//		logStr := fmt.Sprintf("获取进程号失败: [taskId] %v [error] %v", taskId, getProcessIdErr.Error())
		//		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, logStr)
		//		return
		//	}
		//	if processId != "" {
		//		losStr := fmt.Sprintf("正在执行B程序 不启动新的B程序: [taskId] %v [processId] %v", taskId, processId)
		//		logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, losStr)
		//		return
		//	}
		//	_, callSendPublishingErr := callSendPublishing(taskId, headerKey)
		//	if callSendPublishingErr != nil {
		//		logStr := fmt.Sprintf("执行B程序失败: [taskId] %v [error] %v", taskId, callSendPublishingErr.Error())
		//		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, logStr)
		//		return
		//	}
		//	lock.DestroyLock(taskId) //销毁锁
		//}
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
		updateHeaderStatusErr := service.UpdateHeaderStatus(taskId, int64(_type.TaskStatusRunning), expiration)
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
			fmt.Printf("获取BookInfo失败-原因: %v\n", GetTaskBookErr)
			continue
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
		//time.Sleep(time.Millisecond)
		num.Add(1)
		err = service.UpdateTaskCountTrue(taskId, 1, expiration)
	}
	//if err != nil {
	//失败重试一次
	//err = redis.UpdateTaskHeader(redis0, taskId, header, expiration)
	//if err != nil {
	//	headerJson, jsonErr := json.Marshal(header)
	//	if jsonErr != nil {
	//		fmt.Println("JSON编码错误:", err)
	//		return
	//	}
	//	logStr, _ := fmt.Printf("重试更新header任务数量失败: [taskId] %v [header] %v [error] %v", taskId, string(headerJson), err.Error())
	//	logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, strconv.Itoa(logStr))
	//	return
	//}
	//}
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

// callSendPublishing 调用SendPublishing程序处理任务（内部方法）
// 返回进程句柄和错误
func callSendPublishing(qName string, headerKey string) (*os.Process, error) {
	// 先在Redis中创建一个占位符，表示进程即将启动
	placeholderPID := "starting"

	// 同步保存占位符到 Redis
	setProcessIdErr := service.SetProcessId(headerKey, placeholderPID)
	if setProcessIdErr != nil {
		return nil, fmt.Errorf("保存进程占位符到Redis失败: %v, 队列: %s", setProcessIdErr, qName)
	}

	// 构建 SendPublishing程序路径
	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		return nil, getFileUrlConfigErr
	}
	programPath := fileUrlConfig.BFileName

	// 修正参数顺序，去掉-NoExit（如果不是PowerShell程序）
	psScript := fmt.Sprintf(`
        # 启动程序
        $psi = New-Object System.Diagnostics.ProcessStartInfo
        $psi.FileName = "%s"
        $psi.Arguments = "%s"
        $psi.UseShellExecute = $true
        $psi.WindowStyle = 'Normal'
        
        $process = [System.Diagnostics.Process]::Start($psi)
        
        # 等待窗口句柄可用
        $timeout = 3000  # 3秒超时
        $startTime = Get-Date
        while ($process.MainWindowHandle -eq [IntPtr]::Zero -and 
               ((Get-Date) - $startTime).TotalMilliseconds -lt $timeout) {
            Start-Sleep -Milliseconds 50
            $process.Refresh()
        }
        
        # 如果获取到窗口句柄，提到前台
        if ($process.MainWindowHandle -ne [IntPtr]::Zero) {
            Add-Type @"
                using System;
                using System.Runtime.InteropServices;
                public class WindowHelper {
                    [DllImport("user32.dll")]
                    public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
                    [DllImport("user32.dll")]
                    public static extern bool SetForegroundWindow(IntPtr hWnd);
                    [DllImport("user32.dll")]
                    public static extern bool AllowSetForegroundWindow(uint dwProcessId);
                }
"@
            # 允许设置前台窗口
            [WindowHelper]::AllowSetForegroundWindow($process.Id)
            [WindowHelper]::ShowWindow($process.MainWindowHandle, 9)  # SW_RESTORE
            [WindowHelper]::SetForegroundWindow($process.MainWindowHandle)
        } else {
            Write-Warning "未能获取到窗口句柄，但进程已启动 (PID: $($process.Id))"
        }
        
        Write-Output $process.Id
    `, programPath, qName)
	cmd := exec.Command("powershell", "-Command", psScript)
	// 关键配置：使程序在独立控制台运行
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x00000010, // 创建新控制台
		}
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 解析PID
	var pid uint32
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if pids, err := strconv.Atoi(line); err == nil {
			pid = uint32(pids)
		}
	}

	// 获取进程句柄
	// 立即将真实 PID保存到Redis
	processID := fmt.Sprintf("%d", pid)
	setProcessIdErr = service.SetProcessId(headerKey, processID)
	if setProcessIdErr != nil {
		return nil, fmt.Errorf("更新进程PID到Redis失败: %v, 队列: %s", setProcessIdErr, qName)
	}

	// 启动 goroutine等待进程结束并清理 Redis
	go func(pid int, qName string, headerKey string) {
		// 定期检查进程是否还在运行
		for {
			time.Sleep(2 * time.Second) // 每2秒检查一次
			if !isProcessExistWindows(pid) {
				// 进程已结束，清理Redis
				if headerKey != "" {
					deleteProcessIdErr := service.DeleteProcessId(headerKey)
					if deleteProcessIdErr != nil {
						logStr := fmt.Sprintf("清理Redis进程记录失败，PID: %d, 队列: %s, 错误: %v", pid, qName, deleteProcessIdErr)
						logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, logStr)
					}
					break
				}
			}
		}

	}(int(pid), qName, headerKey)

	//将生成的process和redisKey 存储到文件 executingfile
	return &os.Process{Pid: int(pid)}, nil
}

// callSendPublishing 调用SendPublishing程序处理任务（内部方法）
// 返回进程句柄和错误
func callSendPublishing1(qName string, headerKey string) (*os.Process, error) {
	// 先在Redis中创建一个占位符，表示进程即将启动
	placeholderPID := "starting"

	// 同步保存占位符到 Redis
	setProcessIdErr := service.SetProcessId(headerKey, placeholderPID)
	if setProcessIdErr != nil {
		return nil, fmt.Errorf("保存进程占位符到Redis失败: %v, 队列: %s", setProcessIdErr, qName)
	}

	// 构建 SendPublishing程序路径
	fileUrlConfig, getFileUrlConfigErr := config.GetFileUrlConfig()
	if getFileUrlConfigErr != nil {
		return nil, getFileUrlConfigErr
	}
	programPath := fileUrlConfig.BFileName
	oldWindowName := "无标题 - 记事本" // 原窗口名（用于查找）
	// 构建 PowerShell 脚本（包含修改窗口名逻辑）
	psScript := fmt.Sprintf(`
        # 启动程序
        $psi = New-Object System.Diagnostics.ProcessStartInfo
        $psi.FileName = "%s"
        $psi.Arguments = "%s"
        $psi.UseShellExecute = $true
        $psi.WindowStyle = 'Normal'
        
        $process = [System.Diagnostics.Process]::Start($psi)
        
        # 等待窗口创建（超时5秒，延长超时避免窗口未加载完成）
        $timeout = 5000
        $startTime = Get-Date
        $oldWindowName = "%s"    # 原窗口名（用于查找）
        $newWindowName = "%s"    # 新窗口名（要修改成的名称）
        $hwnd = [IntPtr]::Zero

        # 导入所有需要的Windows API
        Add-Type @"
            using System;
            using System.Runtime.InteropServices;
            using System.Text;
            public class WindowHelper {
                // 查找窗口（根据原窗口名）
                [DllImport("user32.dll", CharSet = CharSet.Unicode)]
                public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
                
                // 获取窗口标题（验证用）
                [DllImport("user32.dll", CharSet = CharSet.Unicode)]
                public static extern int GetWindowText(IntPtr hWnd, StringBuilder lpString, int nMaxCount);
                
                // 修改窗口标题（核心API）
                [DllImport("user32.dll", CharSet = CharSet.Unicode)]
                public static extern bool SetWindowText(IntPtr hWnd, string lpString);
                
                // 显示/恢复窗口
                [DllImport("user32.dll")]
                public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
                
                // 置顶窗口
                [DllImport("user32.dll")]
                public static extern bool SetForegroundWindow(IntPtr hWnd);
                
                // 允许设置前台窗口
                [DllImport("user32.dll")]
                public static extern bool AllowSetForegroundWindow(uint dwProcessId);
            }
"@

        # 循环查找指定窗口（延长轮询时间，确保找到）
        while ($hwnd -eq [IntPtr]::Zero -and ((Get-Date) - $startTime).TotalMilliseconds -lt $timeout) {
            Start-Sleep -Milliseconds 100  # 延长轮询间隔，减少资源占用
            $process.Refresh()
            $hwnd = [WindowHelper]::FindWindow($null, $oldWindowName)
        }
        
        # 找到窗口后执行操作
        if ($hwnd -ne [IntPtr]::Zero) {
            # 1. 修改窗口名称（核心步骤）
            $renameSuccess = [WindowHelper]::SetWindowText($hwnd, $newWindowName)
            if ($renameSuccess) {
                Write-Output "窗口名修改成功：$oldWindowName -> $newWindowName"
            } else {
                Write-Warning "窗口名修改失败（权限/窗口不支持）"
            }

            # 2. 验证修改结果
            $sb = New-Object System.Text.StringBuilder(256)
            [WindowHelper]::GetWindowText($hwnd, $sb, $sb.Capacity)
            Write-Output "当前窗口名：$($sb.ToString())"

            # 3. 激活窗口
            [WindowHelper]::AllowSetForegroundWindow($process.Id)
            [WindowHelper]::ShowWindow($hwnd, 9)  # SW_RESTORE
            [WindowHelper]::SetForegroundWindow($hwnd)
            Write-Output "进程PID: $($process.Id)"
        } else {
            Write-Warning "超时未找到窗口：'$oldWindowName' (进程PID: $($process.Id))"
            # 降级使用进程默认窗口句柄
            if ($process.MainWindowHandle -ne [IntPtr]::Zero) {
                [WindowHelper]::SetWindowText($process.MainWindowHandle, $newWindowName)
                [WindowHelper]::AllowSetForegroundWindow($process.Id)
                [WindowHelper]::ShowWindow($process.MainWindowHandle, 9)
                [WindowHelper]::SetForegroundWindow($process.MainWindowHandle)
                Write-Output "已修改进程默认窗口名：$newWindowName"
            }
        }
    `,
		strings.ReplaceAll(programPath, "\"", "\\\""),
		strings.ReplaceAll(qName, "\"", "\\\""),
		strings.ReplaceAll(oldWindowName, "\"", "\\\""), // 原窗口名
		strings.ReplaceAll(qName, "\"", "\\\""))         // 新窗口名

	// 执行PowerShell脚本
	cmd := exec.Command("powershell", "-Command", psScript)
	// Windows下创建新控制台
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x00000010, // CREATE_NEW_CONSOLE
		}
	}

	// 获取输出（包含错误信息，方便排查）
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// 解析PID
	var pid uint32
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if pids, err := strconv.Atoi(line); err == nil {
			pid = uint32(pids)
		}
	}

	// 获取进程句柄
	// 立即将真实 PID保存到Redis
	processID := fmt.Sprintf("%d", pid)
	setProcessIdErr = service.SetProcessId(headerKey, processID)
	if setProcessIdErr != nil {
		return nil, fmt.Errorf("更新进程PID到Redis失败: %v, 队列: %s", setProcessIdErr, qName)
	}

	// 启动 goroutine等待进程结束并清理 Redis
	go func(pid int, qName string, headerKey string) {
		// 定期检查进程是否还在运行
		for {
			time.Sleep(2 * time.Second) // 每2秒检查一次
			if !isProcessExistWindows(pid) {
				// 进程已结束，清理Redis
				if headerKey != "" {
					deleteProcessIdErr := service.DeleteProcessId(headerKey)
					if deleteProcessIdErr != nil {
						logStr := fmt.Sprintf("清理Redis进程记录失败，PID: %d, 队列: %s, 错误: %v", pid, qName, deleteProcessIdErr)
						logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, logStr)
					}
					break
				}
			}
		}

	}(int(pid), qName, headerKey)

	//将生成的process和redisKey 存储到文件 executingfile
	return &os.Process{Pid: int(pid)}, nil
}

// SuspendProcess 暂停指定PID的进程
func SuspendProcess(pidStr string) error {
	// 将字符串转换为整数
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("PID格式错误: %s, 错误: %v", pidStr, err)
	}

	// 检查PID是否有效
	if pid <= 0 {
		return fmt.Errorf("PID必须为正整数")
	}

	// 打开进程
	hProcess, _, err := procOpenProcess.Call(
		PROCESS_SUSPEND_RESUME,
		uintptr(0),
		uintptr(pid),
	)

	if hProcess == 0 {
		return fmt.Errorf("打开进程失败: %v", err)
	}
	defer procCloseHandle.Call(hProcess)

	// 暂停进程
	status, _, _ := procNtSuspendProcess.Call(hProcess)
	if status != 0 {
		return fmt.Errorf("NtSuspendProcess 失败: 0x%X", status)
	}

	return nil
}

// ResumeProcess 恢复指定PID的进程
func ResumeProcess(pidStr string) error {

	// 将字符串转换为整数
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("PID格式错误: %s, 错误: %v", pidStr, err)
	}

	// 检查PID是否有效
	if pid <= 0 {
		return fmt.Errorf("PID必须为正整数")
	}
	// 打开进程
	hProcess, _, err := procOpenProcess.Call(
		PROCESS_SUSPEND_RESUME,
		uintptr(0),
		uintptr(pid),
	)

	if hProcess == 0 {
		return fmt.Errorf("打开进程失败: %v", err)
	}
	defer procCloseHandle.Call(hProcess)

	// 恢复进程
	status, _, _ := procNtResumeProcess.Call(hProcess)
	if status != 0 {
		return fmt.Errorf("NtResumeProcess 失败: 0x%X", status)
	}

	return nil
}

// IsProcessExist 检查进程是否存在（跨平台）
func IsProcessExist(pid int) bool {
	// FindProcess 总是返回一个 Process 对象，即使进程不存在
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 不同平台的实现
	if runtime.GOOS == "windows" {
		// Windows 特定实现
		return isProcessExistWindows(pid)
	}

	// Unix/Linux/MacOS 实现
	err = process.Signal(os.Interrupt)
	if err != nil {
		// 如果错误是 "process already finished"，表示进程不存在
		return false
	}
	return true
}

// Windows 特定的进程检查
func isProcessExistWindows(pid int) bool {
	// 方法1: 尝试获取进程句柄
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	// 检查进程退出码
	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}

	// 如果退出码是 STILL_ACTIVE，表示进程还在运行
	return exitCode == 259 // 259 = STILL_ACTIVE
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
// @param completed 完成数量
// @return error
func AppendToCSV(fileName string, data []_type.TaskBody, writeHeader bool, taskId string, completed *int) error {

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
		*completed++
		err := service.UpdateExportFileProgress(taskId, *completed)
		if err != nil {
			return fmt.Errorf("更新redis进度失败: %v", err)
		}
	}

	return nil
}

// ExportToCSV 将数据导出到export目录下的指定CSV文件
// @param fileName 文件名
// @param data 数据
// @return error 错误信息
func ExportToCSV(fileName string, data []_type.TaskBody) error {
	// 1. 定义导出目录
	exportDir := "export"

	// 2. 检查并创建目录（如果不存在）
	err := os.MkdirAll(exportDir, 0755) // 0755是目录权限，保证可读可写
	if err != nil {
		return fmt.Errorf("创建export目录失败: %w", err)
	}

	// 3. 拼接完整的文件路径（自动处理路径分隔符，兼容Windows/Linux）
	fullPath := filepath.Join(exportDir, fileName)

	// 4. 创建文件（存在则覆盖）
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("创建文件失败 %s: %w", fullPath, err)
	}
	// 延迟关闭文件
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("关闭文件失败 %s: %v", fullPath, closeErr)
		}
	}()

	// 5. 创建CSV写入器并写入数据
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"ISBN", "书名", "状态", "错误信息"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("写入表头失败: %w", err)
	}

	// 写入数据行
	for _, v := range data {
		statusStr := "正确"
		if v.Detail.Status != 1 {
			statusStr = "错误"
		}
		record := []string{
			v.BookInfo.Isbn,
			v.BookInfo.BookName,
			statusStr,
			v.Detail.Error,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入数据行失败: %w", err)
		}
	}

	// 检查写入器错误
	if err := writer.Error(); err != nil {
		return fmt.Errorf("CSV写入器错误: %w", err)
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
