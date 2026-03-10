package tool

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateUUID 生成一个版本4（随机）的UUID字符串
// 返回值：uuid字符串，错误信息（如果生成失败）
func GenerateUUID() (string, error) {
	// NewUUID 生成版本4的随机UUID（最常用的类型）
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("生成UUID失败: %v", err)
	}
	// 将UUID对象转为字符串（标准格式：8-4-4-4-12）
	return uuidObj.String(), nil
}
