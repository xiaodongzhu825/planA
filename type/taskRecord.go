package _type

import "time"

// TaskRecordsDTO 任务记录
type TaskRecordsDTO struct {
	Id       int64     `json:"id"`
	UserId   string    `json:"user_id"`
	ShopId   string    `json:"shop_id"`
	TaskId   string    `json:"task_id"`
	ShopName string    `json:"shop_name"`
	IsExport int64     `json:"is_export"`
	TaskType int64     `json:"task_type"`
	CreateAt time.Time `json:"create_at"`
}

// GetTaskRecordsListReq 获取任务记录列表
type GetTaskRecordsListReq struct {
	UserId   string `json:"user_id"`
	TaskId   string `json:"task_id"`
	TaskType int64  `json:"task_type"`
	ShopName string `json:"shop_name"`
	Page     int
	Size     int
}

// GetTaskRecordsByTaskId 获取任务记录
type GetTaskRecordsByTaskId struct {
	UserId int64  `json:"user_id"`
	TaskId string `json:"task_id"`
}
