package service

import (
	"encoding/json"
	"fmt"
	"planA/initialization/golabl"
	_type "planA/type"
)

// ============================================
// 书籍信息操作
// 数据结构: String (JSON格式)
// 键格式: {bookKey}
// ============================================

// GetTaskBook 获取书籍信息
// @param bookKey 书籍键
// @return _type.BookInfo 书籍信息
// @return error 错误信息
func GetTaskBook(bookKey string) (_type.BookInfo, error) {
	var book _type.BookInfo

	bookStr, err := golabl.RedisDbB.Get(golabl.Ctx, bookKey).Result()
	if err != nil {
		return book, fmt.Errorf("获取书品信息错误: key=%v err=%w", bookKey, err)
	}

	if err := json.Unmarshal([]byte(bookStr), &book); err != nil {
		return book, fmt.Errorf("JSON解析错误: %w", err)
	}

	return book, nil
}

// GetTaskBookPing 测试书籍信息连接（仅用于心跳检测）
func GetTaskBookPing() {
	golabl.RedisDbB.Get(golabl.Ctx, "ping..")
}
