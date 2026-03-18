package _type

import (
	"database/sql"
	"time"
)

// TaskExportDTO 导出任务结构体
type TaskExportDTO struct {
	Id         int64        `json:"id"`
	UserId     string       `json:"user_id"`
	ShopId     string       `json:"shop_id"`
	TaskId     string       `json:"task_id"`
	ShopName   string       `json:"shop_name"`
	FileUrl    string       `json:"file_url"`
	Status     int64        `json:"status"`
	Total      int64        `json:"total"`
	CompleteAt sql.NullTime `json:"complete_at"`
	CreateAt   time.Time    `json:"create_at"`
}

// GetTaskExportListReq 获取导出任务列表
type GetTaskExportListReq struct {
	UserId int64 `json:"user_id"`
	Page   int
	Size   int
}
