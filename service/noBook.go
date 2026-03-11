package service

import (
	"planA/initialization/golabl"
)

// SetNoBookCount 无书籍信息isbn计次
// @param key 键
// @return error 错误信息
func SetNoBookCount(isbn string) error {
	key := "noBookInfo"
	return golabl.RedisDbD.ZIncrBy(golabl.Ctx, key, 1, isbn).Err()
}
