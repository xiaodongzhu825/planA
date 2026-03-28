package service

import (
	"encoding/json"
	"fmt"
	"planA/planB/initialization/golabl"
	planAType "planA/type"
	"strconv"
)

// GetTaskHeader 获取任务头
// @param header *_type.TaskHeader 任务头
// @return error 错误信息
func GetTaskHeader() error {
	// 测试 client 是否可用
	pingErr := golabl.Redis.RedisDbA.Ping(golabl.Ctx).Err()
	if pingErr != nil {
		return pingErr
	}
	// 拼接 key 值
	headerKey := fmt.Sprintf("%s:header", golabl.Task.TaskId)
	// 获取 header 数据
	headerData, hGetAllErr := golabl.Redis.RedisDbA.HGetAll(golabl.Ctx, headerKey).Result()
	if hGetAllErr != nil {
		return fmt.Errorf("获取 header 失败 %w", hGetAllErr)
	}
	// 判断 headerData 是否为空
	if headerData == nil || len(headerData) == 0 {
		return fmt.Errorf("获取 header 失败 %s", "任务信息为空")
	}
	// 解析 header 数据
	parseTaskHeaderErr := parseTaskHeader(headerData)
	if parseTaskHeaderErr != nil {
		return fmt.Errorf("解析 header 失败 %w", parseTaskHeaderErr)
	}
	// 返回结果
	return nil
}

// GetTaskFooter 获取任务尾
// @param error 错误信息
func GetTaskFooter() error {
	// 测试 client 是否可用
	pingErr := golabl.Redis.RedisDbA.Ping(golabl.Ctx).Err()
	if pingErr != nil {
		return pingErr
	}
	// 拼接 key 值
	footerKey := fmt.Sprintf("%s:footer", golabl.Task.TaskId)
	// 获取 footer 数据
	footerData, HGetAllErr := golabl.Redis.RedisDbA.HGetAll(golabl.Ctx, footerKey).Result()
	if HGetAllErr != nil {
		return fmt.Errorf("获取 footer 失败: %w", HGetAllErr)
	}

	// 解析 footer 数据
	parseTaskFooterErr := parseTaskFooter(footerData, golabl.Task.Footer)
	if parseTaskFooterErr != nil {
		return fmt.Errorf("解析 footer 失败: %w", parseTaskFooterErr)
	}

	// 返回结果
	return nil
}

// UpdateTaskHeaderCount 更新任务头
// @return error 错误信息
func UpdateTaskHeaderCount() error {
	// 测试 client 是否可用
	err := golabl.Redis.RedisDbA.Ping(golabl.Ctx).Err()
	if err != nil {
		return err
	}

	// 检查键是否存在
	exists, existsErr := golabl.Redis.RedisDbA.Exists(golabl.Ctx, golabl.Task.TaskId+":header").Result()
	if existsErr != nil {
		return existsErr
	}

	// 键不存在
	if exists == 0 {
		return fmt.Errorf("任务不存在%v", golabl.Task.TaskId)
	}

	// 使用 Pipeline 逐个字段更新
	pipe := golabl.Redis.RedisDbA.Pipeline()
	pipe.HSet(golabl.Ctx, golabl.Task.TaskId+":header", "task_count_wait", golabl.Task.Header.TaskCountWait)
	pipe.HSet(golabl.Ctx, golabl.Task.TaskId+":header", "task_count_over", golabl.Task.Header.TaskCountOver)
	pipe.HSet(golabl.Ctx, golabl.Task.TaskId+":header", "task_count_success", golabl.Task.Header.TaskCountSuccess)
	pipe.HSet(golabl.Ctx, golabl.Task.TaskId+":header", "task_count_error", golabl.Task.Header.TaskCountError)
	_, ExecErr := pipe.Exec(golabl.Ctx)
	if ExecErr != nil {
		return ExecErr
	}

	// 返回结果
	return nil
}

// UpdateTaskFooter 更新任务尾
// @param returnErr int64 类型 1=正确 2= 错误
// @return error 错误信息
func UpdateTaskFooter(returnErr int64) error {
	// 测试 client 是否可用
	err := golabl.Redis.RedisDbA.Ping(golabl.Ctx).Err()
	if err != nil {
		return err
	}

	// 检查键是否存在
	footerKey := golabl.Task.TaskId + ":footer"
	exists, existsErr := golabl.Redis.RedisDbA.Exists(golabl.Ctx, footerKey).Result()
	if existsErr != nil {
		return existsErr
	}
	// 键不存在
	if exists == 0 {
		return fmt.Errorf("任务不存在%v", golabl.Task.TaskId)
	}

	// 使用 Pipeline 逐个字段更新
	pipe := golabl.Redis.RedisDbA.Pipeline()
	// 更新任务尾
	if returnErr == 1 {
		pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_success", 1)
	} else {
		pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_error", 1)
	}
	pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_wait", -1)
	pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_over", 1)
	_, ExecErr := pipe.Exec(golabl.Ctx)
	if ExecErr != nil {
		return ExecErr
	}

	// 返回结果
	return nil
}

// GetTaskToPopFromBodyWait 获取任务信息
// @return _type.TaskBody 任务信息
// @return error 错误信息
func GetTaskToPopFromBodyWait() (planAType.TaskBody, error) {
	// 测试 client 是否可用
	pingErr := golabl.Redis.RedisDbA.Ping(golabl.Ctx).Err()
	if pingErr != nil {
		return planAType.TaskBody{}, pingErr
	}
	// 获取 body 数据
	bodyData, rPopErr := golabl.Redis.RedisDbA.LPop(golabl.Ctx, golabl.Task.TaskId+":body_wait").Result()
	if rPopErr != nil {
		return planAType.TaskBody{}, rPopErr
	}

	// 判断 body 数据是否为空
	if bodyData == "" {
		return planAType.TaskBody{}, fmt.Errorf("任务详情信息为空")
	}
	// 解析 bodyDetail 数据
	taskBody, parseTaskBodyErr := parseTaskBody(bodyData)
	if parseTaskBodyErr != nil {
		return planAType.TaskBody{}, fmt.Errorf("解析任务详情信息失败: %v\n", parseTaskBodyErr)
	}
	// 判断任务状态
	if taskBody.Detail.Status == 3 {
		return planAType.TaskBody{}, fmt.Errorf("任务已执行完毕\n")
	}

	// 返回结果
	return taskBody, nil
}

// SetNoImgCount 无图片信息isbn计次
// @param isbn
// @return error 错误信息
func SetNoImgCount(isbn string) error {
	key := "noImgInfo"
	return golabl.Redis.RedisDbD.ZIncrBy(golabl.Ctx, key, 1, isbn).Err()
}

// AddTaskToBodyOver 添加任务到完成任务池
// @param taskBody _type.TaskBody 任务信息
// @param addType []string 写入类型 ["body_over","body_data","body_backup"]
// @return error 错误信息
func AddTaskToBodyOver(taskBody planAType.TaskBody, addType []string) error {
	// 测试 client 是否可用
	pingErr := golabl.Redis.RedisDbA.Ping(golabl.Ctx).Err()
	if pingErr != nil {
		return pingErr
	}

	// 序列化任务数据
	taskBodyStr, jsonMarshalErr := json.Marshal(taskBody)
	if jsonMarshalErr != nil {
		return fmt.Errorf("任务信息转换失败: %v\n", jsonMarshalErr)
	}

	// 使用事务确保 LPUSH 操作的原子性
	pipe := golabl.Redis.RedisDbA.TxPipeline()

	// 判断需要写入哪些类型
	// 如果 addType 为空数组，则全部写入
	if len(addType) == 0 {
		// 全部写入
		pipe.LPush(golabl.Ctx, golabl.Task.TaskId+":body_over", taskBodyStr)
		pipe.LPush(golabl.Ctx, golabl.Task.TaskId+":body_data", taskBodyStr)
		pipe.LPush(golabl.Ctx, golabl.Task.TaskId+":body_backup", taskBodyStr)
	} else {
		// 根据传入的类型动态写入
		for _, t := range addType {
			switch t {
			case "body_over":
				pipe.LPush(golabl.Ctx, golabl.Task.TaskId+":body_over", taskBodyStr)
			case "body_data":
				pipe.LPush(golabl.Ctx, golabl.Task.TaskId+":body_data", taskBodyStr)
			case "body_backup":
				pipe.LPush(golabl.Ctx, golabl.Task.TaskId+":body_backup", taskBodyStr)
			default:
				// 忽略未知类型，或者可以根据需要返回错误
				continue
			}
		}
	}

	// 执行事务
	_, execErr := pipe.Exec(golabl.Ctx)
	if execErr != nil {
		return fmt.Errorf("添加任务到完成任务池失败: %v\n", execErr)
	}

	// 返回结果
	return nil
}

// GetTaskBodyWaitCount 获取指定body_wait的真实数量
func GetTaskBodyWaitCount() (int64, error) {
	return golabl.Redis.RedisDbA.LLen(golabl.Ctx, golabl.Task.TaskId+":body_wait").Result()
}

// IsTaskBodyWaitExist 查询body_wait是否存在
func IsTaskBodyWaitExist() (bool, error) {
	count, err := golabl.Redis.RedisDbA.Exists(golabl.Ctx, golabl.Task.TaskId+":body_wait").Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetTaskBodyWaitLast 获取body_wait中最后一条数据
func GetTaskBodyWaitLast() (string, error) {
	return golabl.Redis.RedisDbA.LIndex(golabl.Ctx, golabl.Task.TaskId+":body_wait", 1).Result()
}

// AddTaskToBodyWait 写入到body_wait中
// @param taskBody _type.TaskBody 任务信息
// @return error 错误信息
func AddTaskToBodyWait(bodyWaitJson string) error {
	return golabl.Redis.RedisDbA.LPush(golabl.Ctx, golabl.Task.TaskId+":body_wait", bodyWaitJson).Err()
}

// GetTaskBodyWaitList 读取body_wait 数据
// @param page int 页码
// @param pageSize int 页大小
// @return []string body_wait 数据
// @return error 错误信息
func GetTaskBodyWaitList(page int, pageSize int) ([]string, error) {
	// 计算起始索引和结束索引
	// Redis LRange 使用 0-based 索引
	// 第1页: 0 到 pageSize-1
	// 第2页: pageSize 到 2*pageSize-1
	start := (page - 1) * pageSize
	end := start + pageSize - 1

	// 如果页码小于1，默认为第1页
	if page < 1 {
		page = 1
		start = 0
		end = pageSize - 1
	}

	// 如果页大小小于1，设置默认值
	if pageSize < 1 {
		pageSize = 10
		end = start + pageSize - 1
	}

	// 获取指定范围的数据
	return golabl.Redis.RedisDbA.LRange(golabl.Ctx, golabl.Task.TaskId+":body_wait", int64(start), int64(end)).Result()
}

// DeleteTaskBodyWait 删除body_wait 数据
func DeleteTaskBodyWait() error {
	return golabl.Redis.RedisDbA.Del(golabl.Ctx, golabl.Task.TaskId+":body_wait").Err()
}

// DeleteTaskBodyOver 删除body_over 数据
func DeleteTaskBodyOver() error {
	return golabl.Redis.RedisDbA.Del(golabl.Ctx, golabl.Task.TaskId+":body_over").Err()
}

// DeleteTaskBodyBackup 删除body_backup 数据
func DeleteTaskBodyBackup() error {
	return golabl.Redis.RedisDbA.Del(golabl.Ctx, golabl.Task.TaskId+":body_backup").Err()
}

// =========================== 以下是私有方法 ===========================

// 解析任务头
func parseTaskHeader(taskHeader map[string]string) error {

	// 解析 header task_id
	if golabl.Task.Header.TaskId = taskHeader["task_id"]; golabl.Task.Header.TaskId == "" {
		return fmt.Errorf("参数错误: %s", "task_id 为 空")
	}
	// 解析 header shop_id
	if golabl.Task.Header.ShopId, _ = strconv.ParseInt(taskHeader["shop_id"], 10, 64); golabl.Task.Header.ShopId == 0 {
		return fmt.Errorf("参数错误: %s", "shop_id 为 空")
	}
	// 解析 header shop_name
	if golabl.Task.Header.ShopName, _ = taskHeader["shop_name"]; golabl.Task.Header.ShopName == "" {
		return fmt.Errorf("参数错误: %s", "shop_name 为 空")
	}
	// 解析 header shop_msg
	err := json.Unmarshal([]byte(taskHeader["shop_msg"]), &golabl.Task.Header.ShopMsg)
	if err != nil {
		return fmt.Errorf("参数错误: %s", "shop_msg 转结构体失败 shopMsg:="+taskHeader["shop_msg"])
	}
	// 解析 header price_mod
	err = json.Unmarshal([]byte(taskHeader["price_mod"]), &golabl.Task.Header.PriceMod)
	if err != nil {
		return fmt.Errorf("参数错误: %s", "price_mod 转结构体失败 priceMod:="+taskHeader["price_mod"])
	}

	// 解析 header ship_price_mod
	//if header.ShipPriceMod, _ = taskHeader["ship_price_mod"]; header.ShipPriceMod == "" {
	//	return fmt.Errorf("参数错误: %s", "ship_price_mod 为 空")
	//}
	// 解析 header task_type
	if golabl.Task.Header.TaskType, _ = strconv.ParseInt(taskHeader["task_type"], 10, 64); golabl.Task.Header.TaskType == 0 {
		return fmt.Errorf("参数错误: %s", "task_type 为 空")
	}
	// 解析 header shop_type
	if golabl.Task.Header.ShopType, _ = taskHeader["shop_type"]; golabl.Task.Header.ShopType == "" {
		return fmt.Errorf("参数错误: %s", "shop_type 为 空")
	}
	// 解析 header task_count
	if golabl.Task.Header.TaskCount, _ = strconv.ParseInt(taskHeader["task_count"], 10, 64); golabl.Task.Header.TaskCount == 0 {
		//return fmt.Errorf("参数错误: %s", "task_count 为 空")
	}
	// 解析 header task_count_true
	if golabl.Task.Header.TaskCountTrue, _ = strconv.ParseInt(taskHeader["task_count_true"], 10, 64); golabl.Task.Header.TaskCountTrue == 0 {
		//return fmt.Errorf("参数错误: %s ", "task_count_true 为 空")
	}
	// 解析 header task_count_wait
	if golabl.Task.Header.TaskCountWait, _ = strconv.ParseInt(taskHeader["task_count_wait"], 10, 64); golabl.Task.Header.TaskCountWait == 0 {
		//return fmt.Errorf("参数错误: %s", "task_count_wait 为 空")
	}
	// 解析 header task_count_over
	if golabl.Task.Header.TaskCountOver, _ = strconv.ParseInt(taskHeader["task_count_over"], 10, 64); golabl.Task.Header.TaskCountOver == 0 {
		//return fmt.Errorf("参数错误: %s", "task_count_over 为 空")
	}
	// 解析 header task_count_success
	if golabl.Task.Header.TaskCountSuccess, _ = strconv.ParseInt(taskHeader["task_count_success"], 10, 64); golabl.Task.Header.TaskCountSuccess == 0 {
		//return fmt.Errorf("参数错误: %s", "task_count_success 为 空")
	}
	// 解析 header task_count_error
	if golabl.Task.Header.TaskCountError, _ = strconv.ParseInt(taskHeader["task_count_error"], 10, 64); golabl.Task.Header.TaskCountError == 0 {
		//return fmt.Errorf("参数错误: %s", "task_count_error 为 空")
	}
	// 将headerData["status"] 转换为 TaskStatus
	taskStatus, _ := strconv.ParseInt(taskHeader["status"], 10, 64)
	// 解析 header status
	if golabl.Task.Header.Status = planAType.TaskStatus(taskStatus); golabl.Task.Header.Status == 5 {
		return fmt.Errorf("参数错误: %s", "Status 为 已完成任务")
	}
	// 解析 header task_qpm
	if golabl.Task.Header.TaskQpm, _ = strconv.ParseInt(taskHeader["task_qpm"], 10, 64); golabl.Task.Header.TaskQpm == 0 {
		//return fmt.Errorf("参数错误: %s", "task_qpm 为 空")
	}
	// 解析 header task_create_at
	if golabl.Task.Header.TaskCreateAt, _ = strconv.ParseInt(taskHeader["task_create_at"], 10, 64); golabl.Task.Header.TaskCreateAt == 0 {
		//return fmt.Errorf("参数错误: %s", "task_create_at 为 空")
	}
	// 解析 header task_over_at
	if golabl.Task.Header.TaskOverAt, _ = strconv.ParseInt(taskHeader["task_over_at"], 10, 64); golabl.Task.Header.TaskOverAt == 0 {
		//return fmt.Errorf("参数错误: %s", "task_over_at 为 空")
	}
	// 解析 header last_index
	if golabl.Task.Header.LastIndex, _ = strconv.ParseInt(taskHeader["last_index"], 10, 64); golabl.Task.Header.LastIndex == 0 {
		//return fmt.Errorf("参数错误: %s", "last_index 为 空")
	}
	// 解析 header img_type
	if golabl.Task.Header.ImgType, _ = strconv.ParseInt(taskHeader["img_type"], 10, 64); golabl.Task.Header.ImgType == 0 {
		//return fmt.Errorf("参数错误: %s", "last_index 为 空")
	}
	// 解析 header pool
	if taskHeader["pool"] != "" {
		err = json.Unmarshal([]byte(taskHeader["pool"]), &golabl.Task.Header.Pool)
		if err != nil {
			return fmt.Errorf("参数错误: %s", "pool 转结构体失败 pool:="+taskHeader["pool"])
		}
	} else {
		// 空字符串时，初始化为空的切片或结构体
		golabl.Task.Header.Pool = planAType.PoolConfig{} // 如果是切片类型
	}

	// 返回结果
	return nil
}

// 解析任务尾
func parseTaskFooter(taskFooter map[string]string, footer *planAType.TaskFooter) error {
	// 解析 footer task_count
	if footer.TaskCount, _ = strconv.ParseInt(taskFooter["task_count"], 10, 64); footer.TaskCount == 0 {
	}
	// 解析 footer task_count_true
	if footer.TaskCountTrue, _ = strconv.ParseInt(taskFooter["task_count_true"], 10, 64); footer.TaskCountTrue == 0 {
	}
	// 解析 footer task_count_wait
	taskCountWait, _ := strconv.ParseInt(taskFooter["task_count_wait"], 10, 64)
	if taskCountWait == 0 {
	}
	footer.TaskCountWait.Store(taskCountWait)
	// 解析 footer task_count_over
	taskCountOver, _ := strconv.ParseInt(taskFooter["task_count_over"], 10, 64)
	if taskCountOver == 0 {
	}
	footer.TaskCountOver.Store(taskCountOver)
	// 解析 footer task_count_success
	taskCountSuccess, _ := strconv.ParseInt(taskFooter["task_count_success"], 10, 64)
	if taskCountSuccess == 0 {
	}
	footer.TaskCountSuccess.Store(taskCountSuccess)
	// 解析 footer task_count_error
	taskCountError, _ := strconv.ParseInt(taskFooter["task_count_error"], 10, 64)
	if taskCountError == 0 {
	}
	footer.TaskCountError.Store(taskCountError)
	// 解析 footer task_qpm
	if footer.TaskQpm, _ = strconv.ParseInt(taskFooter["task_qpm"], 10, 64); footer.TaskQpm == 0 {
	}
	// 解析 footer last_index
	if footer.LastIndex, _ = strconv.ParseInt(taskFooter["last_index"], 10, 64); footer.LastIndex == 0 {
	}

	// 返回结果
	return nil
}

// 解析任务主体
func parseTaskBody(taskBodyStr string) (planAType.TaskBody, error) {
	// 初始化 body
	var body planAType.TaskBody
	// 解析 bookInfo 到 结构体
	UnmarshalErr := json.Unmarshal([]byte(taskBodyStr), &body)
	if UnmarshalErr != nil {
		return planAType.TaskBody{}, UnmarshalErr
	}

	// 返回结果
	return body, nil
}
