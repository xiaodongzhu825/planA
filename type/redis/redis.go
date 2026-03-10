package redis

import "encoding/json"

// RedisData 原始Redis数据结构
type RedisData struct {
	SourceTable string          `json:"source_table"`
	Data        json.RawMessage `json:"data"`
}
