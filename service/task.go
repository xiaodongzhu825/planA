package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/initialization/golabl"
	"planA/tool"
	_type "planA/type"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// ============================================
// 任务头信息(Header)操作
// 数据结构: Hash
// 键格式: {taskKey}:header
// ============================================

// GetTaskHeader 获取任务头信息
// @param client Redis客户端
// @param taskKey 任务键
// @return _type.TaskHeader 任务头信息
// @return error 错误信息
func GetTaskHeader(taskKey string) (_type.TaskHeader, error) {
	var header _type.TaskHeader
	headerKey := getHeaderKey(taskKey)
	headerMap, err := golabl.RedisDbA.HGetAll(golabl.Ctx, headerKey).Result()
	if err != nil {
		return header, err
	}

	return parseHeaderMap(headerMap)
}

// UpdateTaskHeader 更新任务头信息
// @param taskKey 任务键
// @param header 任务头信息
// @param expiration 过期时间
// @return error 错误信息
func UpdateTaskHeader(taskKey string, header _type.TaskHeader, expiration time.Duration) error {
	// 将结构体转为 map
	headerMap, err := tool.StructToMap(header)
	if err != nil {
		return fmt.Errorf("转换Header为map失败: %w", err)
	}

	// 特殊处理price_mod字段
	priceModJSON, err := json.Marshal(headerMap["price_mod"])
	if err != nil {
		return fmt.Errorf("转换price_mod为JSON失败: %w", err)
	}
	headerMap["price_mod"] = priceModJSON
	// 保存到 Redis
	headerKey := getHeaderKey(taskKey)
	if err := saveHashMap(headerKey, headerMap, expiration); err != nil {
		return err
	}
	return nil
}

// UpdateHeaderStatus 更新任务头信息中的状态
// @param client Redis客户端
// @param taskKey 任务键
// @param status 任务状态（1=运行中 2=已暂停 3=已停止）
// @param expiration 过期时间
// @return error 错误信息
func UpdateHeaderStatus(taskKey string, status int64, expiration time.Duration) error {
	headerKey := getHeaderKey(taskKey)
	if err := golabl.RedisDbA.HSet(golabl.Ctx, headerKey, "status", status).Err(); err != nil {
		return err
	}
	golabl.RedisDbA.Expire(golabl.Ctx, headerKey, expiration)
	return nil
}

// ============================================
// 任务体信息(Body)操作
// 数据结构: List
// 键格式: {taskKey}:body_wait - 等待处理的任务队列
//        {taskKey}:body_over  - 已完成处理的任务队列
// ============================================

// UpdateTaskBodyWait 添加任务到等待队列
// @param taskKey 任务键
// @param taskBody 任务体数据
// @return error 错误信息
func UpdateTaskBodyWait(taskKey string, taskBody _type.TaskBody) error {
	bodyWaitKey := getBodyWaitKey(taskKey)

	// 序列化任务数据
	bodyWaitJSON, err := json.Marshal(taskBody)
	if err != nil {
		return fmt.Errorf("序列化任务数据失败: %w", err)
	}

	// 推送到列表尾部
	return golabl.RedisDbA.RPush(golabl.Ctx, bodyWaitKey, string(bodyWaitJSON)).Err()
}

// GetListLength 获取等待队列长度
// @param taskKey 任务键
// @return int64 队列长度
// @return error 错误信息
func GetListLength(taskKey string) (int64, error) {
	bodyWaitKey := getBodyWaitKey(taskKey)
	return golabl.RedisDbA.LLen(golabl.Ctx, bodyWaitKey).Result()
}

// GetTaskBodyOverLimit10 获取最近10条已完成任务
// @param taskKey 任务键
// @return []_type.TaskBody 任务体列表
// @return error 错误信息
func GetTaskBodyOverLimit10(taskKey string) ([]_type.TaskBody, error) {
	return GetBodyOverDataByBatch(taskKey, 0, 10)
}

// GetBodyOverCount 获取已完成任务总数
// @param taskKey 任务键
// @return int64 总数
// @return error 错误信息
func GetBodyOverCount(taskKey string) (int64, error) {
	bodyOverKey := getBodyOverKey(taskKey)
	return golabl.RedisDbA.LLen(golabl.Ctx, bodyOverKey).Result()
}

// GetBodyOverDataByBatch 批量获取已完成任务数据
// @param taskKey 任务键
// @param offset 起始偏移量
// @param count 获取数量
// @return []_type.TaskBody 任务体列表
// @return error 错误信息
func GetBodyOverDataByBatch(taskKey string, offset, count int) ([]_type.TaskBody, error) {
	var bodyOverArr []_type.TaskBody

	bodyOverKey := getBodyOverKey(taskKey)
	end := offset + count - 1 // LRange是闭区间 [start, end]

	bodyOverStr, err := golabl.RedisDbA.LRange(golabl.Ctx, bodyOverKey, int64(offset), int64(end)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return bodyOverArr, nil
		}
		return bodyOverArr, fmt.Errorf("获取body_over数据错误: %w", err)
	}

	return parseTaskBodyList(bodyOverStr)
}

// ClearBodyOver 清空已完成任务队列
// @param client Redis客户端
// @param taskKey 任务键
// @return error 错误信息
func ClearBodyOver(taskKey string) error {
	return golabl.RedisDbA.Del(golabl.Ctx, getBodyOverKey(taskKey)).Err()
}

// ============================================
// 任务尾信息(Footer)操作
// 数据结构: Hash
// 键格式: {taskKey}:footer
// ============================================

// GetTaskFooter 获取任务尾信息
// @param taskKey 任务键
// @return _type.TaskFooter 任务尾信息
// @return error 错误信息
func GetTaskFooter(taskKey string) (_type.TaskFooter, error) {
	var footer _type.TaskFooter
	footerKey := getFooterKey(taskKey)
	footerMap, err := golabl.RedisDbA.HGetAll(golabl.Ctx, footerKey).Result()
	if err != nil {
		return footer, fmt.Errorf("获取Footer失败: %w", err)
	}

	return footer, parseTaskFooter(footerMap, &footer)
}

// UpdateTaskFooter 更新任务尾信息
// @param taskKey 任务键
// @param footer 任务尾信息
// @param expiration 过期时间
// @return error 错误信息
func UpdateTaskFooter(taskKey string, footer *_type.TaskFooter, expiration time.Duration) error {
	footerMap := map[string]interface{}{
		"task_count":         footer.TaskCount,
		"task_count_true":    footer.TaskCountTrue,
		"task_count_wait":    footer.TaskCountWait.Load(),
		"task_count_over":    footer.TaskCountOver.Load(),
		"task_count_success": footer.TaskCountSuccess.Load(),
		"task_count_error":   footer.TaskCountError.Load(),
		"task_qpm":           footer.TaskQpm,
		"last_index":         footer.LastIndex,
	}

	footerKey := getFooterKey(taskKey)
	if err := saveHashMap(footerKey, footerMap, expiration); err != nil {
		return err
	}

	return nil
}

// ============================================
// 导出文件(BodyFile)操作
// 数据结构: Hash
// 键格式: {taskKey}:body_file
// ============================================

// UpdateExportFileProgress 更新导出文件进度
// @param taskKey 任务键
// @param complete 完成进度
// @return error 错误信息
func UpdateExportFileProgress(taskKey string, complete int) error {
	return golabl.RedisDbA.HSet(golabl.Ctx, getBodyFileKey(taskKey), "complete", complete).Err()
}

// GetExportFileProgress 获取导出文件进度
// @param taskKey 任务键
// @return int 完成进度
// @return error 错误信息
func GetExportFileProgress(taskKey string) (int, error) {
	return golabl.RedisDbA.HGet(golabl.Ctx, getBodyFileKey(taskKey), "complete").Int()
}

// ============================================
// 进程号管理操作
// 数据结构: Hash字段
// 键格式: {headerKey}
// 字段: process_number
// ============================================

// GetProcessId 获取进程号
// @param headerKey 头信息键
// @return string 进程号
// @return error 错误信息
func GetProcessId(headerKey string) (string, error) {
	return golabl.RedisDbA.HGet(golabl.Ctx, headerKey, "process_number").Result()
}

// SetProcessId 设置进程号
// @param headerKey 头信息键
// @param processId 进程号
// @return error 错误信息
func SetProcessId(headerKey string, processId string) error {
	return golabl.RedisDbA.HSet(golabl.Ctx, headerKey, "process_number", processId).Err()
}

// DeleteProcessId 删除进程号
// @param headerKey 头信息键
// @return error 错误信息
func DeleteProcessId(headerKey string) error {
	return golabl.RedisDbA.HDel(golabl.Ctx, headerKey, "process_number").Err()
}

// ============================================
// 复合操作（涉及多个数据结构）
// ============================================

// UpdateTaskCountTrue 更新任务计数（原子操作）
// 同时更新Header和Footer中的task_count_true，以及Footer中的task_count_wait
// @param taskKey 任务键
// @param num 增减数量
// @param expiration 过期时间
// @return error 错误信息
func UpdateTaskCountTrue(taskKey string, num int64, expiration time.Duration) error {
	// 使用Pipeline确保原子性
	pipe := golabl.RedisDbA.Pipeline()

	// 更新Header
	headerKey := getHeaderKey(taskKey)
	pipe.HIncrBy(golabl.Ctx, headerKey, "task_count_true", num)

	// 更新Footer
	footerKey := getFooterKey(taskKey)
	pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_true", num)
	pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_wait", num)

	// 设置过期时间
	if expiration > 0 {
		pipe.Expire(golabl.Ctx, headerKey, expiration)
		pipe.Expire(golabl.Ctx, footerKey, expiration)
	}

	return executePipeline(pipe)
}

// StopTask 停止任务
// 更新Header状态为已停止，并清空等待队列
// @param taskId 任务ID
// @param expiration 过期时间
// @return error 错误信息
func StopTask(taskId string, expiration time.Duration) error {
	// 开启事务
	pipe := golabl.RedisDbA.TxPipeline()

	// 更新任务状态
	headerKey := getHeaderKey(taskId)
	pipe.HSet(golabl.Ctx, headerKey, "status", int64(_type.TaskStatusStopped))

	// 设置过期时间
	if expiration > 0 {
		pipe.Expire(golabl.Ctx, headerKey, expiration)
	}

	// 清空等待队列
	bodyWaitKey := getBodyWaitKey(taskId)
	pipe.Del(golabl.Ctx, bodyWaitKey)

	return executePipeline(pipe)
}

//********************************************以下为是有方法*****************************************//

// getHeaderKey 获取任务头信息的Redis键
// @param taskKey 任务键
// @return string 头信息键
func getHeaderKey(taskKey string) string {
	return taskKey + ":header"
}

// getFooterKey 获取任务尾信息的Redis键
// @param taskKey 任务键
// @return string 尾信息键
func getFooterKey(taskKey string) string {
	return taskKey + ":footer"
}

// getBodyWaitKey 获取等待任务体的Redis键
// @param taskKey 任务键
// @return string 等待任务体键
func getBodyWaitKey(taskKey string) string {
	return taskKey + ":body_wait"
}

// getBodyOverKey 获取已完成任务体的Redis键
// @param taskKey 任务键
// @return string 已完成任务体键
func getBodyOverKey(taskKey string) string {
	return taskKey + ":body_over"
}

// getBodyDataKey 获取已完成任务体的Redis键
// @param taskKey 任务键
// @return string 已完成任务体键
func getBodyDataKey(taskKey string) string {
	return taskKey + ":body_data"
}

// getBodyBackupKey 获取已完成任务体的Redis键
// @param taskKey 任务键
// @return string 已完成任务体键
func getBodyBackupKey(taskKey string) string {
	return taskKey + ":body_backup"
}

// getBodyFileKey 获取已完成任务体的Redis键
// @param taskKey 任务键
// @return string 已完成任务体键
func getBodyFileKey(taskKey string) string {
	return taskKey + ":body_file"
}

// parseHeaderMap 从map解析任务头信息
// @param headerMap 头信息map
// @return _type.TaskHeader 解析后的头信息
// @return error 错误信息
func parseHeaderMap(headerMap map[string]string) (_type.TaskHeader, error) {
	info := _type.TaskHeader{}

	for key, value := range headerMap {
		switch key {
		case "last_index", "shop_id", "task_count", "task_count_error",
			"task_count_over", "task_count_success", "task_count_true",
			"task_count_wait", "task_create_at", "task_over_at", "task_qpm", "task_type":
			parseIntField(&info, key, value)

		case "price_mod":
			var priceMod []_type.PriceMod
			if err := json.Unmarshal([]byte(value), &priceMod); err == nil {
				info.PriceMod = priceMod
			}

		case "shop_msg":
			var shopMsg _type.ShopMsg
			if err := json.Unmarshal([]byte(value), &shopMsg); err == nil {
				info.ShopMsg = shopMsg
			}

		case "status":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				info.Status = _type.TaskStatus(v)
			}

		case "ship_price_mod", "shop_name", "shop_type", "task_id":
			setStringField(&info, key, value)
		}
	}
	return info, nil
}

// parseIntField 解析整数字段
func parseIntField(info *_type.TaskHeader, key, value string) {
	if v, err := strconv.ParseInt(value, 10, 64); err == nil {
		switch key {
		case "last_index":
			info.LastIndex = v
		case "shop_id":
			info.ShopId = v
		case "task_count":
			info.TaskCount = v
		case "task_count_error":
			info.TaskCountError = v
		case "task_count_over":
			info.TaskCountOver = v
		case "task_count_success":
			info.TaskCountSuccess = v
		case "task_count_true":
			info.TaskCountTrue = v
		case "task_count_wait":
			info.TaskCountWait = v
		case "task_create_at":
			info.TaskCreateAt = v
		case "task_over_at":
			info.TaskOverAt = v
		case "task_qpm":
			info.TaskQpm = v
		case "task_type":
			info.TaskType = v
		}
	}
}

// setStringField 设置字符串字段
func setStringField(info *_type.TaskHeader, key, value string) {
	switch key {
	case "ship_price_mod":
		info.ShipPriceMod = value
	case "shop_name":
		info.ShopName = value
	case "shop_type":
		info.ShopType = value
	case "task_id":
		info.TaskId = value
	}
}

// saveHashMap 保存哈希映射到Redis
// @param key Redis键
// @param data 数据映射
// @return error 错误信息
func saveHashMap(key string, data map[string]interface{}, expiration time.Duration) error {
	for field, value := range data {
		if err := golabl.RedisDbA.HSet(golabl.Ctx, key, field, value).Err(); err != nil {
			return fmt.Errorf("保存字段 %s 失败: %w (值: %v)", field, err, value)
		}
	}
	golabl.RedisDbA.Expire(golabl.Ctx, key, expiration)
	return nil
}

// parseTaskBodyList 解析任务体列表
// @param bodyStrs 任务体字符串列表
// @return []_type.TaskBody 解析后的任务体列表
// @return error 错误信息
func parseTaskBodyList(bodyStrs []string) ([]_type.TaskBody, error) {
	var bodyList []_type.TaskBody

	for _, str := range bodyStrs {
		var body _type.TaskBody
		if err := json.Unmarshal([]byte(str), &body); err != nil {
			return bodyList, fmt.Errorf("JSON解析错误: %w, 数据: %s", err, str)
		}
		bodyList = append(bodyList, body)
	}

	return bodyList, nil
}

// parseTaskFooter 解析任务尾信息
// @param taskFooter 尾信息map
// @param footer 目标尾信息结构体
// @return error 错误信息
func parseTaskFooter(taskFooter map[string]string, footer *_type.TaskFooter) error {
	var err error

	if footer.TaskCount, err = strconv.ParseInt(taskFooter["task_count"], 10, 64); err != nil {
		footer.TaskCount = 0
	}

	if footer.TaskCountTrue, err = strconv.ParseInt(taskFooter["task_count_true"], 10, 64); err != nil {
		footer.TaskCountTrue = 0
	}

	if taskCountWait, err := strconv.ParseInt(taskFooter["task_count_wait"], 10, 64); err == nil {
		footer.TaskCountWait.Store(taskCountWait)
	}

	if taskCountOver, err := strconv.ParseInt(taskFooter["task_count_over"], 10, 64); err == nil {
		footer.TaskCountOver.Store(taskCountOver)
	}

	if taskCountSuccess, err := strconv.ParseInt(taskFooter["task_count_success"], 10, 64); err == nil {
		footer.TaskCountSuccess.Store(taskCountSuccess)
	}

	if taskCountError, err := strconv.ParseInt(taskFooter["task_count_error"], 10, 64); err == nil {
		footer.TaskCountError.Store(taskCountError)
	}

	if footer.TaskQpm, err = strconv.ParseInt(taskFooter["task_qpm"], 10, 64); err != nil {
		footer.TaskQpm = 0
	}

	if footer.LastIndex, err = strconv.ParseInt(taskFooter["last_index"], 10, 64); err != nil {
		footer.LastIndex = 0
	}

	return nil
}

// executePipeline 执行Redis管道操作
// @param pipe Redis管道
// @return error 错误信息
func executePipeline(pipe redis.Pipeliner) error {
	_, err := pipe.Exec(golabl.Ctx)
	return err
}
