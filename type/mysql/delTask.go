package mysql

import (
	"time"

	"gorm.io/gorm"
)

// DelTask 删除任务表
// 对应数据库中的 del_task 表
type DelTask struct {
	// ID 主键，自增
	ID int64 `gorm:"column:id;type:int(11);primaryKey;autoIncrement;comment:主键ID" json:"id"`

	// UserID 用户ID
	UserID *string `gorm:"column:user_id;type:varchar(64);index:idx_user_shop_task;comment:用户ID" json:"user_id,omitempty"`

	// ShopID 店铺ID
	ShopID *int64 `gorm:"column:shop_id;type:bigint(64);index:idx_user_shop_task;comment:店铺ID" json:"shop_id,omitempty"`

	// TaskID 任务ID
	TaskID *string `gorm:"column:task_id;type:varchar(64);index:idx_user_shop_task;comment:任务ID" json:"task_id,omitempty"`

	// pid pid
	Pid *string `gorm:"column:pid;type:varchar(64);index:idx_user_shop_task;comment:pid" json:"pid,omitempty"`

	// ShopName 店铺名称
	ShopName *string `gorm:"column:shop_name;type:varchar(128);index:idx_user_shop_task;comment:店铺名称" json:"shop_name,omitempty"`

	// TaskCount 任务数
	TaskCount *int `gorm:"column:task_count;type:int(11);index:idx_user_shop_task;comment:任务数" json:"task_count,omitempty"`

	// TaskCountOver 完成任务数
	TaskCountOver *int `gorm:"column:task_count_over;type:int(11);index:idx_user_shop_task;comment:完成任务数" json:"task_count_over,omitempty"`

	// Status 状态
	Status *int `gorm:"column:status;type:int(11);index:idx_user_shop_task;comment:状态 1=执行中 2=暂停 3=完成" json:"status,omitempty"`

	// header 任务头信息
	Header *string `gorm:"column:header;type:text;comment:任务头" json:"header,omitempty"`

	// pauseAt 暂停时间
	PauseAt *time.Time `gorm:"column:pause_at;type:datetime;comment:暂停时间" json:"pause_at"`

	// CreateAt 创建时间（GORM会自动维护创建时间）
	CreateAt *time.Time `gorm:"column:create_at;type:datetime;autoCreateTime;comment:创建时间" json:"create_at,omitempty"`
}

// TableName 指定结构体对应的数据库表名
func (t *DelTask) TableName() string {
	return "del_task"
}

// MigrateDelTask 初始化表结构/索引
// @param db 数据库连接实例
// @return error 错误信息
func MigrateDelTask(db *gorm.DB) error {
	return db.AutoMigrate(&DelTask{})
}
