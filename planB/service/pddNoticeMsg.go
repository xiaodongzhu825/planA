package service

import "planA/planB/initialization/golabl"

// GetPddNoticeMsg 获取拼多多通知消息
// @param shopId 店铺ID
// @return []string 消息列表
// @return error 错误
func GetPddNoticeMsg(shopId string) ([]string, error) {
	// 测试 client 是否可用
	pingErr := golabl.Redis.RedisDbA.Ping(golabl.Ctx).Err()
	if pingErr != nil {
		return []string{}, pingErr
	}
	//获取所有list数据
	list, lRangeErr := golabl.Redis.RedisDbA.LRange(golabl.Ctx, shopId, 0, -1).Result()
	if lRangeErr != nil {
		return []string{}, lRangeErr
	}
	return list, nil
}
