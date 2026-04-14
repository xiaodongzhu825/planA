package _type

// GoodsType 接口类型
type GoodsType string

// AsyncTaskResponse 添加商品数据返回结构体
type AsyncTaskResponse struct {
	Msg             string `json:"msg"`             // 消息说明
	CurrentProgress int    `json:"currentProgress"` // 当前进度
	Code            string `json:"code"`            // 状态码
	TaskKey         string `json:"taskKey"`         // 任务唯一标识
	TotalCount      int    `json:"totalCount"`      // 总数量
}
