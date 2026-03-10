package sqLite

import (
	"database/sql"
	"time"
)

// TaskExport 定义任务导出表(task_export)对应的结构体
// 该表用于存储任务导出的详细信息
type TaskExport struct {
	ID         int          // 主键 ID
	UserID     int64        // 用户 ID
	TaskID     string       // 任务 ID
	ShopName   string       // 店铺名称
	FileUrl    string       // 导出文件 URL
	Status     int          // 状态(0:未开始 1:进行中 2:完成)
	Total      int          // 总数量
	CompleteAt sql.NullTime // 完成时间
	CreateAt   time.Time    // 创建时间
}
