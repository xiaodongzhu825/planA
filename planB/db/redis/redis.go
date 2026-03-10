package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"planA/planB/tool"
	_type "planA/planB/type"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	gCtx context.Context
)

// Init 初始化
func Init(ctx context.Context, config _type.RedisConfig) (*redis.Client, error) {
	// 判断 ctx 是否取消
	checkContextErr := tool.CheckContext(ctx)
	// 判断 结果
	if checkContextErr != nil {
		// 返回 且 返回错误
		return nil, checkContextErr
	}
	// 设置全局 ctx
	gCtx = ctx

	// 初始化 redis
	return redis.NewClient(&redis.Options{
		Addr:               config.Addr,                              // 连接地址
		Password:           config.Password,                          // 密码
		DB:                 config.DB,                                // 数据库
		PoolSize:           config.PoolSize,                          // 连接池大小
		PoolTimeout:        time.Duration(config.PoolTimeout),        // 连接池超时时间
		ReadTimeout:        time.Duration(config.ReadTimeout),        // 读取超时
		WriteTimeout:       time.Duration(config.WriteTimeout),       // 写入超时
		DialTimeout:        time.Duration(config.DialTimeout),        // 连接超时
		IdleTimeout:        time.Duration(config.IdleTimeout),        // 空闲超时
		MinIdleConns:       config.MinIdleConns,                      // 最小空闲连接数
		IdleCheckFrequency: time.Duration(config.IdleCheckFrequency), // 空闲检查频率
		MaxRetries:         config.MaxRetries,                        // 最大重试次数
		MaxRetryBackoff:    time.Duration(config.MaxRetryBackoff),    // 最大重试间隔
		MinRetryBackoff:    time.Duration(config.MinRetryBackoff),    // 最小重试间隔
	}), nil
}

// GetTaskHeader 获取任务头
func GetTaskHeader(client *redis.Client, taskKey string, header *_type.TaskHeader) error {
	// 测试 client 是否可用
	pingErr := client.Ping(gCtx).Err()
	if pingErr != nil {
		return pingErr
	}
	// 拼接 key 值
	headerKey := fmt.Sprintf("%s:header", taskKey)
	// 获取 header 数据
	headerData, hGetAllErr := client.HGetAll(gCtx, headerKey).Result()
	if hGetAllErr != nil {
		return fmt.Errorf("获取 header 失败: %w", hGetAllErr)
	}
	//fmt.Printf("headerData 为:%v", headerData)
	// 判断 headerData 是否为空
	if headerData == nil || len(headerData) == 0 {
		return fmt.Errorf("获取 header 失败: %s", "任务信息为空")
	}
	// 解析 header 数据
	parseTaskHeaderErr := parseTaskHeader(headerData, header)
	if parseTaskHeaderErr != nil {
		return fmt.Errorf("解析 header 失败: %w", parseTaskHeaderErr)
	}
	// 返回结果
	return nil
}

// GetTaskFooter 获取任务尾
func GetTaskFooter(client *redis.Client, taskKey string, taskFooter *_type.TaskFooter) error {
	// 测试 client 是否可用
	pingErr := client.Ping(gCtx).Err()
	if pingErr != nil {
		return pingErr
	}
	// 拼接 key 值
	footerKey := fmt.Sprintf("%s:footer", taskKey)
	// 获取 footer 数据
	footerData, HGetAllErr := client.HGetAll(gCtx, footerKey).Result()
	if HGetAllErr != nil {
		return fmt.Errorf("获取 footer 失败: %w", HGetAllErr)
	}

	// 解析 footer 数据
	parseTaskFooterErr := parseTaskFooter(footerData, taskFooter)
	if parseTaskFooterErr != nil {
		return fmt.Errorf("解析 footer 失败: %w", parseTaskFooterErr)
	}

	// 返回结果
	return nil
}

// GetTaskToPopFromBodyWait 获取任务信息
func GetTaskToPopFromBodyWait(client *redis.Client, taskKey string) (_type.TaskBody, error) {
	// 测试 client 是否可用
	pingErr := client.Ping(gCtx).Err()
	if pingErr != nil {
		return _type.TaskBody{}, pingErr
	}
	// 获取 body 数据
	bodyData, rPopErr := client.LPop(gCtx, taskKey+":body_wait").Result() // 正式环境使用
	//bodyData, rPopErr := client.LIndex(gCtx, taskKey+":body_wait", 0).Result() // 测试环境使用
	if rPopErr != nil {
		return _type.TaskBody{}, fmt.Errorf("读取任务详情信息失败: %v\n", rPopErr)
	}

	// 判断 body 数据是否为空
	if bodyData == "" {
		return _type.TaskBody{}, fmt.Errorf("任务详情信息为空")
	}
	// 解析 bodyDetail 数据
	taskBody, parseTaskBodyErr := parseTaskBody(bodyData)
	if parseTaskBodyErr != nil {
		return _type.TaskBody{}, fmt.Errorf("解析任务详情信息失败: %v\n", parseTaskBodyErr)
	}
	// 判断任务状态
	if taskBody.Detail.Status == 3 {
		return _type.TaskBody{}, fmt.Errorf("任务已执行完毕\n")
	}

	// 返回结果
	return taskBody, nil
}

// UpdateTaskHeaderCount 更新任务头
func UpdateTaskHeaderCount(client *redis.Client, taskKey string, taskHeader _type.TaskHeader) (bool, error) {
	// 测试 client 是否可用
	err := client.Ping(gCtx).Err()
	if err != nil {
		return false, err
	}

	// 检查键是否存在
	exists, existsErr := client.Exists(gCtx, taskKey+":header").Result()
	if existsErr != nil {
		return false, existsErr
	}

	// 键不存在
	if exists == 0 {
		return false, fmt.Errorf("任务不存在%v", taskKey)
	}

	// 使用 Pipeline 逐个字段更新
	pipe := client.Pipeline()
	pipe.HSet(gCtx, taskKey+":header", "task_count_wait", taskHeader.TaskCountWait)
	pipe.HSet(gCtx, taskKey+":header", "task_count_over", taskHeader.TaskCountOver)
	pipe.HSet(gCtx, taskKey+":header", "task_count_success", taskHeader.TaskCountSuccess)
	pipe.HSet(gCtx, taskKey+":header", "task_count_error", taskHeader.TaskCountError)
	_, ExecErr := pipe.Exec(gCtx)
	if ExecErr != nil {
		return false, ExecErr
	}

	// 返回结果
	return true, nil
}

// UpdateTaskFooter 更新任务尾
func UpdateTaskFooter(client *redis.Client, taskKey string, taskFooter *_type.TaskFooter, returnErr int64) (bool, error) {
	// 测试 client 是否可用
	err := client.Ping(gCtx).Err()
	if err != nil {
		return false, err
	}

	// 检查键是否存在
	footerKey := taskKey + ":footer"
	exists, existsErr := client.Exists(gCtx, footerKey).Result()
	if existsErr != nil {
		return false, existsErr
	}
	// 键不存在
	if exists == 0 {
		return false, fmt.Errorf("任务不存在%v", taskKey)
	}

	// 使用 Pipeline 逐个字段更新
	pipe := client.Pipeline()
	// 更新任务尾
	if returnErr == 1 {
		pipe.HIncrBy(gCtx, footerKey, "task_count_success", 1)
	} else {
		pipe.HIncrBy(gCtx, footerKey, "task_count_error", 1)
	}
	pipe.HIncrBy(gCtx, footerKey, "task_count_wait", -1)
	pipe.HIncrBy(gCtx, footerKey, "task_count_over", 1)
	_, ExecErr := pipe.Exec(gCtx)
	if ExecErr != nil {
		return false, ExecErr
	}

	// 返回结果
	return true, nil
}

// UpdateTaskMsg 更新任务信息
func UpdateTaskMsg(handle *redis.Client, taskKey string, index int64, err error) (any, error) {
	//TODO

	return nil, nil
}

// AddTaskToBodyOver 添加任务到完成任务池
func AddTaskToBodyOver(client *redis.Client, taskKey string, taskBody _type.TaskBody) (bool, error) {
	// 测试 client 是否可用
	pingErr := client.Ping(gCtx).Err()
	if pingErr != nil {
		return false, pingErr
	}

	// 序列化任务数据
	taskBodyStr, jsonMarshalErr := json.Marshal(taskBody)
	if jsonMarshalErr != nil {
		return false, fmt.Errorf("任务信息转换失败: %v\n", jsonMarshalErr)
	}

	// 使用事务确保两个 LPUSH 操作的原子性
	pipe := client.TxPipeline()

	// 添加body_over任务
	pipe.LPush(gCtx, taskKey+":body_over", taskBodyStr)
	// 添加body_data任务
	pipe.LPush(gCtx, taskKey+":body_data", taskBodyStr)
	// 添加body_data任务
	pipe.LPush(gCtx, taskKey+":body_backup", taskBodyStr)

	// 执行事务
	_, execErr := pipe.Exec(gCtx)
	if execErr != nil {
		return false, fmt.Errorf("添加任务到完成任务池失败: %v\n", execErr)
	}

	// 返回结果
	return true, nil
}

// GetPublishingVid 获取出版社信息Vid
// @param client Redis客户端
// @param publishingName 出版社名称
// @return _type.Publishing 出版社信息
func GetPublishingVid(client *redis.Client, publishingName string) (_type.Publishing, error) {
	var publishing _type.Publishing
	// 测试 client 是否可用
	//pingErr := client.Ping(gCtx).Err()
	//if pingErr != nil {
	//	return publishing, pingErr
	//}
	//获取出版社信息
	publishingStr, getErr := client.Get(gCtx, "publisher:name:"+publishingName).Result()
	if getErr != nil {
		// 出版社不存在，给个默认的
		if errors.Is(getErr, redis.Nil) {
			publishing.Value = "北京大学出版社"
			publishing.Vid = 483727
			return publishing, nil
		}
		return publishing, getErr
	}
	//转为结构体
	unmarshalErr := json.Unmarshal([]byte(publishingStr), &publishing)
	if unmarshalErr != nil {
		return publishing, fmt.Errorf("出版社json转结构体失败 %v", unmarshalErr)
	}
	return publishing, nil
}

// SetShopGoodsRelationData 写入店铺商品关系数据
func SetShopGoodsRelationData(client *redis.Client, shopId int64, isbn string) {

}

// GetRandomDistrict 随机获取一个区级地区
func GetRandomDistrict(client *redis.Client) (map[string]string, error) {
	// 从所有区级地区集合中随机获取一个 ID
	districtID, err := client.SRandMember(gCtx, "all:districts").Result()
	if err != nil {
		return nil, err
	}

	// 获取该地区的详细信息
	return client.HGetAll(gCtx, fmt.Sprintf("region:%s", districtID)).Result()
}

// GetRandomDistrictInProvince 在指定省内随机获取一个区级地区
func GetRandomDistrictInProvince(client *redis.Client, provinceID int) (map[string]string, error) {
	// 从该省份的区级地区集合中随机获取一个 ID
	provinceKey := fmt.Sprintf("province:%d:districts", provinceID)
	districtID, err := client.SRandMember(gCtx, provinceKey).Result()
	if err != nil {
		return nil, err
	}

	// 获取该地区的详细信息
	return client.HGetAll(gCtx, fmt.Sprintf("region:%s", districtID)).Result()
}

// GetProvinceAndCity 根据地区获取获取省市信息
func GetProvinceAndCity(client *redis.Client, districtID int) (int, int, error) {
	// 获取该地区的详细信息
	district, err := client.HGet(gCtx, fmt.Sprintf("region:%s", districtID), "pid").Result()
	fmt.Println("------------------------------")
	fmt.Println(district)
	fmt.Println(fmt.Sprintf("region:%v", districtID))
	fmt.Println("------------------------------")
	if err != nil {
		return 0, 0, err
	}
	//获取省份ID
	//province, err := client.Get(gCtx, fmt.Sprintf("region:%s", district["pid"])).Result()
	//if err != nil {
	//	return 0, 0, err
	//}
	//将 province["id"] 与 district["pid"]  转为 int
	//provinceId, atoiErr := strconv.Atoi(province["id"])
	//if atoiErr != nil {
	//	fmt.Println(province)
	//	fmt.Printf("province_id 转 int 失败: %v %v\n", districtID, atoiErr)
	//	return 0, 0, atoiErr
	//}
	//cityId, atoiErr := strconv.Atoi(district["pid"])
	//if atoiErr != nil {
	//	fmt.Printf("cityId 转 int 失败: %v %v\n", cityId, atoiErr)
	//	return 0, 0, atoiErr
	//}
	return 0, 0, nil
}

// =========================== 以下是私有方法 ===========================

// 解析任务头
func parseTaskHeader(taskHeader map[string]string, header *_type.TaskHeader) error {
	// 解析 header task_id
	if header.TaskId, _ = taskHeader["task_id"]; header.TaskId == "" {
		return fmt.Errorf("参数错误: %s", "task_id 为 空")
	}
	// 解析 header shop_id
	if header.ShopId, _ = strconv.ParseInt(taskHeader["shop_id"], 10, 64); header.ShopId == 0 {
		return fmt.Errorf("参数错误: %s", "shop_id 为 空")
	}
	// 解析 header shop_name
	if header.ShopName, _ = taskHeader["shop_name"]; header.ShopName == "" {
		return fmt.Errorf("参数错误: %s", "shop_name 为 空")
	}
	// 解析 header shop_msg
	err := json.Unmarshal([]byte(taskHeader["shop_msg"]), &header.ShopMsg)
	if err != nil {
		return fmt.Errorf("参数错误: %s", "shop_msg 转结构体失败 shopMsg:="+taskHeader["shop_msg"])
	}
	// 解析 header price_mod
	err = json.Unmarshal([]byte(taskHeader["price_mod"]), &header.PriceMod)
	if err != nil {
		return fmt.Errorf("参数错误: %s", "price_mod 转结构体失败 priceMod:="+taskHeader["price_mod"])
	}

	// 解析 header ship_price_mod
	//if header.ShipPriceMod, _ = taskHeader["ship_price_mod"]; header.ShipPriceMod == "" {
	//	return fmt.Errorf("参数错误: %s", "ship_price_mod 为 空")
	//}
	// 解析 header task_type
	if header.TaskType, _ = strconv.ParseInt(taskHeader["task_type"], 10, 64); header.TaskType == 0 {
		return fmt.Errorf("参数错误: %s", "task_type 为 空")
	}
	// 解析 header shop_type
	if header.ShopType, _ = taskHeader["shop_type"]; header.ShopType == "" {
		return fmt.Errorf("参数错误: %s", "shop_type 为 空")
	}
	// 解析 header task_count
	if header.TaskCount, _ = strconv.ParseInt(taskHeader["task_count"], 10, 64); header.TaskCount == 0 {
		return fmt.Errorf("参数错误: %s", "task_count 为 空")
	}
	// 解析 header task_count_true
	if header.TaskCountTrue, _ = strconv.ParseInt(taskHeader["task_count_true"], 10, 64); header.TaskCountTrue == 0 {
		return fmt.Errorf("参数错误: %s ", "task_count_true 为 空")
	}
	// 解析 header task_count_wait
	if header.TaskCountWait, _ = strconv.ParseInt(taskHeader["task_count_wait"], 10, 64); header.TaskCountWait == 0 {
	}
	// 解析 header task_count_over
	if header.TaskCountOver, _ = strconv.ParseInt(taskHeader["task_count_over"], 10, 64); header.TaskCountOver == 0 {
	}
	// 解析 header task_count_success
	if header.TaskCountSuccess, _ = strconv.ParseInt(taskHeader["task_count_success"], 10, 64); header.TaskCountSuccess == 0 {
	}
	// 解析 header task_count_error
	if header.TaskCountError, _ = strconv.ParseInt(taskHeader["task_count_error"], 10, 64); header.TaskCountError == 0 {
	}
	// 将headerData["status"] 转换为 TaskStatus
	taskStatus, _ := strconv.ParseInt(taskHeader["status"], 10, 64)
	// 解析 header status
	if header.Status = _type.TaskStatus(taskStatus); header.Status == 5 {
		return fmt.Errorf("参数错误: %s", "Status 为 已完成任务")
	}
	// 解析 header task_qpm
	if header.TaskQpm, _ = strconv.ParseInt(taskHeader["task_qpm"], 10, 64); header.TaskQpm == 0 {
	}
	// 解析 header task_create_at
	if header.TaskCreateAt, _ = strconv.ParseInt(taskHeader["task_create_at"], 10, 64); header.TaskCreateAt == 0 {
		return fmt.Errorf("参数错误: %s", "task_create_at 为 空")
	}
	// 解析 header task_over_at
	if header.TaskOverAt, _ = strconv.ParseInt(taskHeader["task_over_at"], 10, 64); header.TaskOverAt == 0 {
	}
	// 解析 header last_index
	if header.LastIndex, _ = strconv.ParseInt(taskHeader["last_index"], 10, 64); header.LastIndex == 0 {
	}

	// 返回结果
	return nil
}

// 解析任务尾
func parseTaskFooter(taskFooter map[string]string, footer *_type.TaskFooter) error {
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
func parseTaskBody(taskBodyStr string) (_type.TaskBody, error) {
	//// 打印 taskBodyJson
	//fmt.Println("taskBodyJson: ", taskBodyStr)

	// 初始化 body
	var body _type.TaskBody
	// 解析 bookInfo 到 结构体
	UnmarshalErr := json.Unmarshal([]byte(taskBodyStr), &body)
	if UnmarshalErr != nil {
		return _type.TaskBody{}, UnmarshalErr
	}

	// 返回结果
	return body, nil
}
