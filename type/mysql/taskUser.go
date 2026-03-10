package mysql

import (
	_type "planA/type"
	"time"

	"gorm.io/gorm"
)

// PageQueryTaskUserByUserIdParams 分页查询参数
type PageQueryTaskUserByUserIdParams struct {
	UserID   int64  // 要查询的用户ID（必传）
	ShopName string // 店铺名称（可选，非空则过滤）
	TaskID   string // 任务ID（可选，非空则过滤）
	TaskType int64  // 任务类型（可选，非空则过滤）
	Page     _type.Page
}

// TaskUser 任务-用户关联表
// 对应数据库中的 task_user 表
type TaskUser struct {
	// ID 主键，自增
	ID int64 `gorm:"column:id;type:int(11);primaryKey;autoIncrement" json:"id"`
	// UserID 用户ID
	UserID *int64 `gorm:"column:user_id;type:bigint(20);index:idx_user_shop_task" json:"user_id,omitempty"`
	// ShopID 店铺ID
	ShopID *int64 `gorm:"column:shop_id;type:bigint(20);index:idx_user_shop_task" json:"shop_id,omitempty"`
	// TaskID 任务ID
	TaskID *string `gorm:"column:task_id;type:varchar(32);index:idx_user_shop_task" json:"task_id,omitempty"`
	// ShopName 店铺名称
	ShopName *string `gorm:"column:shop_name;type:varchar(128);index:idx_user_shop_task" json:"shop_name,omitempty"`
	// IsExport 是否导出，默认为false(0)
	IsExport *int64 `gorm:"column:is_export;type:tinyint(1);default:0" json:"is_export,omitempty"`
	// taskType 是否导出，默认为 1 核价发布 2 表格发布
	TaskType *int64 `gorm:"column:task_type;type:tinyint(1);default:0" json:"task_type,omitempty"`
	// CreateAt 创建时间（GORM会自动维护创建时间）
	CreateAt *time.Time `gorm:"column:create_at;type:datetime;autoCreateTime" json:"create_at,omitempty"`
}

// TableName 指定结构体对应的数据库表名
func (t *TaskUser) TableName() string {
	return "task_user"
}

// MigrateTaskUser 初始化表结构/索引
// @param db 数据库连接实例
// @return error 错误信息
func MigrateTaskUser(db *gorm.DB) error {
	return db.AutoMigrate(&TaskUser{})
}
