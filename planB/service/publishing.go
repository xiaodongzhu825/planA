package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/planB/initialization/golabl"
	planAType "planA/type"

	"github.com/go-redis/redis/v8"
)

// GetPublishingVid 获取出版社信息Vid
// @param taskMsg 任务信息
// @return _type.Publishing 出版社信息
func GetPublishingVid(taskMsg *planAType.TaskBody) error {
	var publishing planAType.Publishing
	//获取出版社信息
	publishingStr, getErr := golabl.Redis.RedisDbB.Get(golabl.Ctx, "publisher:name:"+taskMsg.BookInfo.Publishing).Result()
	if getErr != nil {
		// 出版社不存在，给个默认的
		if errors.Is(getErr, redis.Nil) {
			publishing.Value = "北京大学出版社"
			publishing.Vid = 483727
			return nil
		}
		return getErr
	}
	//转为结构体
	unmarshalErr := json.Unmarshal([]byte(publishingStr), &publishing)
	if unmarshalErr != nil {
		return fmt.Errorf("出版社json转结构体失败 %v", unmarshalErr)
	}
	taskMsg.Publishing = publishing
	return nil
}
