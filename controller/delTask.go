package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	planBType "planA/planB/type"
	"planA/service"
	serviceMysql "planA/service/mysql"
	"planA/tool"
	toolPdd "planA/tool/pdd"
	_type "planA/type"
	mysqlType "planA/type/mysql"
	"planA/validator"
	"strconv"
	"time"
)

// GetDelTaskByPage 分页查询 删除任务
func GetDelTaskByPage(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetDelTaskValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	delTaskArr, total, getDelTaskByPageErr := serviceMysql.GetDelTaskByPage(page, size, "")
	if getDelTaskByPageErr != nil {
		return
	}

	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  delTaskArr,
	}
	tool.Success(httpMsg, dataRet)
	return
}

// GetDelTaskByPageByUserId 分页查询 删除任务-用户
func GetDelTaskByPageByUserId(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetDelTaskByUserIdValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	delTaskArr, total, getDelTaskByPageErr := serviceMysql.GetDelTaskByPage(page, size, dataVal.UserId)
	if getDelTaskByPageErr != nil {
		return
	}

	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  delTaskArr,
	}
	tool.Success(httpMsg, dataRet)
	return
}

// GetDelTaskDetail 获取删除任务详情列表
func GetDelTaskDetail(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetDelTaskDetailValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	delTaskArr, total, getDelTaskByPageErr := serviceMysql.GetDelTaskDetailByPage(page, size, dataVal.TaskId, 3)
	if getDelTaskByPageErr != nil {
		return
	}

	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  delTaskArr,
	}
	tool.Success(httpMsg, dataRet)

}

// CreateTbDelTask 创建淘宝删除任务
func CreateTbDelTask(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.CreateTbDelTaskValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	delNum := data.FormValue("del_num")
	taskCount := 0
	if dataVal.TaskType == "1" || dataVal.TaskType == "2" {
		if delNum == "" || delNum == "0" {
			errMsg := "删除数量不能为空与0"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}

		// 将 dataVal.TaskCount 转为 int
		var err error
		taskCount, err = strconv.Atoi(delNum)
		if err != nil {
			errMsg := "任务数量转换失败: " + err.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	}
	delTime := data.FormValue("del_time")
	delTimes := time.Time{}
	if dataVal.TaskType == "3" {
		if delTime == "" || delTime == "0" {
			errMsg := "删除时间不能为空"
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}

		// 将 dataVal.DelTime 转为 time.Time
		layout := "2006-01-02 15:04:05" // 常见格式
		var err error
		delTimes, err = time.Parse(layout, delTime)
		if err != nil {
			errMsg := "时间转换失败: " + err.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
	}

	// 将 dataVal.TaskType 转为 int
	taskType, err := strconv.Atoi(dataVal.TaskType)
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
	shopData, err := toolPdd.ParseShopData(shopDataStr)
	if err != nil {
		errMsg := "解析店铺数据失败:" + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	if shopData.Shop == nil {
		errMsg := "店铺数据为空"
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	//请求创建任务接口并获取任务 id
	taskId, err := CreateTaskRequest(dataVal.ShopID, dataVal.TaskType)
	if err != nil {
		errMsg := "店铺ID " + dataVal.ShopID + " 淘宝删除任务 请求创建任务接口失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	taskTypes := 0
	if taskType == 1 {
		taskTypes = 5
	} else if taskType == 2 {
		taskTypes = 10
	} else if taskType == 3 {
		taskTypes = 11
	}

	var priceRange []_type.PriceRange
	priceTemplateRangeStr := shopData.PriceTemplate.RangePrice
	err = json.Unmarshal([]byte(priceTemplateRangeStr), &priceRange)
	if err != nil {
		errMsg := "解析价格模板失败:" + err.Error() + " 原始数据：" + shopData.PriceTemplate.RangePrice
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	createAt := time.Now().Unix()
	taskData, createTaskDataErr := CreateTaskData(taskId, int64(taskTypes), createAt, shopData.Shop, priceRange, shopData.Spec, shopData.ShopDetail, shopData.ShopContext, strconv.Itoa(taskCount), 1, 1, "1")
	if createTaskDataErr != nil {
		errMsg := "创建任务数据失败: " + createTaskDataErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 将 taskData 转为json
	taskDataJson, err := json.Marshal(taskData)
	if err != nil {
		errMsg := "任务数据转换失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	header := string(taskDataJson)

	createAts := time.Unix(createAt, 0)
	status := 0
	taskCountOver := 0
	// 创建任务
	task := mysqlType.DelTask{
		UserID:        &shopData.Shop.CreateBy,
		ShopID:        &dataVal.ShopID,
		TaskID:        &taskId,
		ShopType:      &dataVal.TaskType,
		TaskType:      &taskType,
		ShopName:      &shopData.Shop.ShopName,
		TaskCount:     &taskCount,
		Header:        &header,
		CreateAt:      &createAts,
		StopAt:        &delTimes,
		Status:        &status,
		TaskCountOver: &taskCountOver,
	}
	createDelTaskErr := serviceMysql.CreateDelTask(task)
	if createDelTaskErr != nil {
		errMsg := "创建任务失败: " + createDelTaskErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 创建详情表
	createTableIfNotExistsErr := serviceMysql.CreateTableIfNotExists(taskId)
	if createTableIfNotExistsErr != nil {
		errMsg := "创建详情表失败: " + createTableIfNotExistsErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	tool.Success(httpMsg, taskId)
}

// CreateTbDelTaskDetails 插入淘宝删除任务数据
func CreateTbDelTaskDetails(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, createTaskValidatorErr := validator.CreateTbDelTaskDetailsValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	if dataVal.Status == "" {
		dataVal.Status = "0"
	}

	//判断任务是否存在
	_, getTaskErr := serviceMysql.GetDelTaskByTaskId(dataVal.TaskID)
	if getTaskErr != nil {
		if errors.Is(getTaskErr, sql.ErrNoRows) || getTaskErr.Error() == "record not found" {
			errMsg := "任务不存在: " + getTaskErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		errMsg := "获取任务失败: " + getTaskErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 创建详情表
	createTableIfNotExistsErr := serviceMysql.CreateTableIfNotExists(dataVal.TaskID)
	if createTableIfNotExistsErr != nil {
		errMsg := "创建详情表失败: " + createTableIfNotExistsErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 将 string 类型的 Status 转换为 int64
	var statusInt64 int64
	if dataVal.Status != "" {
		statusInt64, _ = strconv.ParseInt(dataVal.Status, 10, 64)
	}

	// 转换 GoodsId
	var goodsIdInt64 int64
	if dataVal.GoodsId != "" {
		if val, err := strconv.ParseInt(dataVal.GoodsId, 10, 64); err == nil {
			goodsIdInt64 = val
		}
	}

	// 转换 TaskID
	var taskIDInt64 int64
	if dataVal.TaskID != "" {
		if val, err := strconv.ParseInt(dataVal.TaskID, 10, 64); err == nil {
			taskIDInt64 = val
		}
	}

	// 获取当前时间
	createAtTime := time.Unix(time.Now().Unix(), 0)
	createAtStr := time.Now().Format("2006-01-02")

	// 插入任务详情
	taskDetails := planBType.DelTaskDetail{
		TaskID:     &dataVal.TaskID,
		Isbn:       &dataVal.Isbn,
		BookName:   &dataVal.BookName,
		GoodsID:    &goodsIdInt64,
		Status:     &statusInt64,
		Err:        &dataVal.Err,
		DeleteAt:   &createAtTime,
		DeleteDate: &createAtStr,
		CreateAt:   &createAtTime,
	}
	insertTbDelTaskDetailsErr := serviceMysql.InsertDelTaskDetail(taskIDInt64, taskDetails)
	if insertTbDelTaskDetailsErr != nil {
		errMsg := "插入任务详情失败: " + insertTbDelTaskDetailsErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	tool.Success(httpMsg, "")
}

// UpdateTbDelTaskDetailsStatus 修改指定淘宝删除任务详情状态
func UpdateTbDelTaskDetailsStatus(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, createTaskValidatorErr := validator.UpdateTbDelTaskDetailsStatusValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	// 判断任务是否存在
	_, getTaskErr := serviceMysql.GetDelTaskByTaskId(dataVal.TaskID)
	if getTaskErr != nil {
		if errors.Is(getTaskErr, sql.ErrNoRows) || getTaskErr.Error() == "record not found" {
			errMsg := "任务不存在: " + getTaskErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		errMsg := "获取任务失败: " + getTaskErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 将 dataVal.Status 转为int
	statusInt, err := strconv.Atoi(dataVal.Status)
	if err != nil {
		errMsg := "状态值格式错误"
		tool.Error(httpMsg, errMsg, http.StatusBadRequest)
		return
	}

	updateTbDelTaskDetailsStatusErr := serviceMysql.UpdateDelTaskDetailStatus(dataVal.TaskID, dataVal.GoodsId, statusInt, dataVal.Err)
	if updateTbDelTaskDetailsStatusErr != nil {
		errMsg := "修改任务详情状态失败: " + updateTbDelTaskDetailsStatusErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	tool.Success(httpMsg, "")
}

// UpdateTbDelTaskProgress 修改淘宝任务进度
func UpdateTbDelTaskProgress(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, createTaskValidatorErr := validator.UpdateTbDelTaskProgressValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	// 判断任务是否存在
	_, getTaskErr := serviceMysql.GetDelTaskByTaskId(dataVal.TaskID)
	if getTaskErr != nil {
		if errors.Is(getTaskErr, sql.ErrNoRows) || getTaskErr.Error() == "record not found" {
			errMsg := "任务不存在: " + getTaskErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		errMsg := "获取任务失败: " + getTaskErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 将 dataVal.Num 转为 int
	numInt, err := strconv.Atoi(dataVal.Num)
	if err != nil {
		errMsg := "进度值格式错误"
		tool.Error(httpMsg, errMsg, http.StatusBadRequest)
		return
	}

	updateTbDelTaskProgressErr := serviceMysql.UpdateDelTaskProgress(dataVal.TaskID, numInt)
	if updateTbDelTaskProgressErr != nil {
		errMsg := "修改任务进度失败: " + updateTbDelTaskProgressErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	tool.Success(httpMsg, "")
}

// UpdateTbDelTaskStatus 修改淘宝任务状态
func UpdateTbDelTaskStatus(httpMsg http.ResponseWriter, data *http.Request) {

	// 验证表单
	dataVal, createTaskValidatorErr := validator.UpdateTbDelTaskStatusValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	// 判断任务是否存在
	_, getTaskErr := serviceMysql.GetDelTaskByTaskId(dataVal.TaskID)
	if getTaskErr != nil {
		if errors.Is(getTaskErr, sql.ErrNoRows) || getTaskErr.Error() == "record not found" {
			errMsg := "任务不存在: " + getTaskErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		errMsg := "获取任务失败: " + getTaskErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 将 dataVal.Status 转为 int
	statusInt, err := strconv.Atoi(dataVal.Status)
	if err != nil {
		errMsg := "状态值格式错误"
		tool.Error(httpMsg, errMsg, http.StatusBadRequest)
		return
	}

	updateTbDelTaskStatusErr := serviceMysql.UpdateDelTaskStatusByTaskId(dataVal.TaskID, statusInt)
	if updateTbDelTaskStatusErr != nil {
		errMsg := "修改任务状态失败: " + updateTbDelTaskStatusErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	tool.Success(httpMsg, "")

}

// GetTbDelTaskDetailsWait 获取指定任务待执行的任务详情
func GetTbDelTaskDetailsWait(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetDelTaskDetailValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	delTaskArr, total, getDelTaskByPageErr := serviceMysql.GetDelTaskDetailByPage(page, size, dataVal.TaskId, 0)
	if getDelTaskByPageErr != nil {
		return
	}

	dataArr := []map[string]interface{}{}
	for _, v := range delTaskArr {
		datas := map[string]interface{}{
			"task_id":   v.TaskID,
			"isbn":      v.Isbn,
			"book_name": v.BookName,
			"goods_id":  v.GoodsID,
			"status":    v.Status,
			"create_at": v.CreateAt,
		}
		dataArr = append(dataArr, datas)
	}

	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  dataArr,
	}
	tool.Success(httpMsg, dataRet)
}

// GetTbDelTaskByTaskId 根据任务id 查询任务
func GetTbDelTaskByTaskId(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.TaskIdValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	// 判断任务是否存在
	task, getTaskErr := serviceMysql.GetDelTaskByTaskId(dataVal.TaskID)
	if getTaskErr != nil {
		if errors.Is(getTaskErr, sql.ErrNoRows) || getTaskErr.Error() == "record not found" {
			errMsg := "任务不存在: " + getTaskErr.Error()
			tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
			return
		}
		errMsg := "获取任务失败: " + getTaskErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	datas := map[string]interface{}{
		"task_id":         task.TaskID,
		"shop_id":         task.ShopID,
		"task_type":       task.TaskType,
		"status":          task.Status,
		"task_count":      task.TaskCount,
		"task_count_over": task.TaskCountOver,
		"stop_at":         task.StopAt,
		"create_at":       task.CreateAt,
	}

	tool.Success(httpMsg, datas)
}
