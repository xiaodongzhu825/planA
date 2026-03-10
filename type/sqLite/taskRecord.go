package sqLite

import "time"

// TaskRecord 定义任务记录表(task_records)对应的结构体
// 该表用于存储任务的基本信息
type TaskRecord struct {
	ID       int       // 主键ID
	UserID   int64     // 用户ID
	TaskID   string    // 任务ID
	ShopName string    // 店铺名称
	IsExport int       // 是否已导出(0:未导出 1:已导出)
	TaskType int64     // 任务类型(1:核价发布 2:表格发布
	CreateAt time.Time // 创建时间
}
