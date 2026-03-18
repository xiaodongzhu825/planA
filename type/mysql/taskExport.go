package mysql

import (
	"database/sql"
	_type "planA/type"
	"time"

	"gorm.io/gorm"
)

// PageQueryTaskExportParams 分页查询参数
type PageQueryTaskExportParams struct {
	UserID int64
	Page   _type.Page
}

// TaskExport
// 对应数据库中的 task_export 表
type TaskExport struct {
	// ID 主键，自增
	ID int64 `gorm:"column:id;type:int(11);primaryKey;autoIncrement;comment:主键ID" json:"id"`

	// UserID 用户ID
	UserID *string `gorm:"column:user_id;type:varchar(64);index:idx_user_shop_task;comment:用户ID" json:"user_id,omitempty"`

	// ShopID 店铺ID
	ShopID *string `gorm:"column:shop_id;type:varchar(64);index:idx_user_shop_task;comment:店铺ID" json:"shop_id,omitempty"`

	// TaskID 任务ID
	TaskID *string `gorm:"column:task_id;type:varchar(64);index:idx_user_shop_task;comment:任务ID" json:"task_id,omitempty"`

	// ShopName 店铺名称
	ShopName *string `gorm:"column:shop_name;type:varchar(128);index:idx_user_shop_task;comment:店铺名称" json:"shop_name,omitempty"`

	// FileUrl 导出文件 URL
	FileUrl *string `gorm:"column:file_url;type:varchar(256);comment:导出文件URL" json:"file_url"`

	// Status 状态(0:未开始 1:进行中 2:完成)
	Status *int64 `gorm:"column:status;type:tinyint(1);default:0;comment:状态(0:未开始 1:进行中 2:完成)" json:"status,omitempty"`

	// Total 总数量
	Total *int64 `gorm:"column:total;type:int(11);comment:总数量" json:"total,omitempty"`

	// CompleteAt 完成时间
	CompleteAt *sql.NullTime `gorm:"column:complete_at;type:datetime;comment:完成时间" json:"complete_at"`

	// CreateAt 创建时间（GORM会自动维护创建时间）
	CreateAt *time.Time `gorm:"column:create_at;type:datetime;autoCreateTime;comment:创建时间" json:"create_at,omitempty"`
}

// TableName 指定结构体对应的数据库表名
func (t *TaskExport) TableName() string {
	return "task_export"
}

// MigrateTaskExport 初始化表结构/索引
// @param db 数据库连接实例
// @return error 错误信息
func MigrateTaskExport(db *gorm.DB) error {
	return db.AutoMigrate(&TaskExport{})
}
