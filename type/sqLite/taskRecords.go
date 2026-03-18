package sqLite

import (
	_type "planA/type"
	"time"
)

// TaskRecords 定义任务记录表(task_records)对应的结构体
// 该表用于存储任务的基本信息
type TaskRecords struct {
	ID       int64     // 主键 ID
	UserID   string    // 用户 ID
	ShopID   string    // 店铺 ID
	TaskID   string    // 任务 ID
	ShopName string    // 店铺名称
	IsExport int64     // 是否已导出(0:未导出 1:已导出)
	TaskType int64     // 任务类型(1:核价发布 2:表格发布
	CreateAt time.Time // 创建时间
}

// GetTaskRecordsByUserIdParams 分页查询参数
type GetTaskRecordsByUserIdParams struct {
	UserID   string // 要查询的用户ID（必传）
	ShopName string // 店铺名称（可选，非空则过滤）
	TaskID   string // 任务ID（可选，非空则过滤）
	TaskType int64  // 任务类型（可选，非空则过滤）
	Page     _type.Page
}
