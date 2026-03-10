package service

import (
	"encoding/json"
	"fmt"
	"planA/initialization/golabl"
	"strings"
)

// ============================================
// 店铺信息操作
// 数据结构: 支持String/List/Hash多种类型
// 键格式: {shopID}
// ============================================

// GetTaskShop 获取店铺信息
// @param shopID 店铺ID
// @return string 店铺信息字符串
// @return error 错误信息
func GetTaskShop(shopID string) (string, error) {
	// 检查键类型
	keyType, err := golabl.RedisDbC.Type(golabl.Ctx, shopID).Result()
	if err != nil {
		return "", fmt.Errorf("检查Redis key类型失败: %w", err)
	}
	switch keyType {
	case "string":
		return golabl.RedisDbC.Get(golabl.Ctx, shopID).Result()

	case "list":
		items, err := golabl.RedisDbC.LRange(golabl.Ctx, shopID, 0, -1).Result()
		if err != nil {
			return "", fmt.Errorf("获取list数据失败: %w", err)
		}
		return "[" + strings.Join(items, ",") + "]", nil

	case "hash":
		hashData, err := golabl.RedisDbC.HGetAll(golabl.Ctx, shopID).Result()
		if err != nil {
			return "", fmt.Errorf("获取hash数据失败: %w", err)
		}
		jsonData, _ := json.Marshal(hashData)
		return string(jsonData), nil

	default:
		return "", fmt.Errorf("不支持的数据类型: %s", keyType)
	}
}
