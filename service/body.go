package service

import (
	"fmt"
	"planA/initialization/golabl"
)

// GetOneBodyFromRight 从右侧获取一条body数据
// 从Redis list中获取最后一条数据并删除原数据
// @param taskId 任务id
// @return string 获取的数据
// @return error 错误信息
func GetOneBodyFromRight(taskId string) (string, error) {
	key := taskId + ":body_wait"
	// RPop: 从列表右侧获取并删除最后一个元素（原子操作）
	data, err := golabl.RedisDbA.RPop(golabl.Ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("从Redis获取body数据失败: %w", err)
	}

	return data, nil
}

// InsertOneBodyOver 插入一条bodyOver数据
func InsertOneBodyOver(taskId string, value string) error {
	key := taskId + ":body_over"
	return golabl.RedisDbA.LPush(golabl.Ctx, key, value).Err()
}

// UpdateTaskFooters 更新任务尾
// @param returnErr int64 类型 1=正确 2= 错误
// @param count int64 类型 更新的数据
// @return error 错误信息
func UpdateTaskFooters(taskId string, returnErr int64, count int64) error {
	// 测试 client 是否可用
	err := golabl.RedisDbA.Ping(golabl.Ctx).Err()
	if err != nil {
		return err
	}

	// 检查键是否存在
	footerKey := taskId + ":footer"
	exists, existsErr := golabl.RedisDbA.Exists(golabl.Ctx, footerKey).Result()
	if existsErr != nil {
		return existsErr
	}
	// 键不存在
	if exists == 0 {
		return fmt.Errorf("任务不存在%v", taskId)
	}

	// 使用 Pipeline 逐个字段更新
	pipe := golabl.RedisDbA.Pipeline()
	// 更新任务尾
	if returnErr == 1 {
		pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_success", count)
	} else {
		pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_error", count)
	}
	pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_wait", -count)
	pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_over", count)
	_, ExecErr := pipe.Exec(golabl.Ctx)
	if ExecErr != nil {
		return ExecErr
	}
	// 返回结果
	return nil
}

// UpdateTaskHeaders 更新任务尾
// @param returnErr int64 类型 1=正确 2= 错误
// @param count int64 类型 更新的数据
// @return error 错误信息
func UpdateTaskHeaders(taskId string, returnErr int64, count int64) error {
	// 测试 client 是否可用
	err := golabl.RedisDbA.Ping(golabl.Ctx).Err()
	if err != nil {
		return err
	}

	// 检查键是否存在
	footerKey := taskId + ":header"
	exists, existsErr := golabl.RedisDbA.Exists(golabl.Ctx, footerKey).Result()
	if existsErr != nil {
		return existsErr
	}
	// 键不存在
	if exists == 0 {
		return fmt.Errorf("任务不存在%v", taskId)
	}

	// 使用 Pipeline 逐个字段更新
	pipe := golabl.RedisDbA.Pipeline()
	// 更新任务尾
	if returnErr == 1 {
		pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_success", count)
	} else {
		pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_error", count)
	}
	pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_wait", -count)
	pipe.HIncrBy(golabl.Ctx, footerKey, "task_count_over", count)
	_, ExecErr := pipe.Exec(golabl.Ctx)
	if ExecErr != nil {
		return ExecErr
	}
	// 返回结果
	return nil
}
