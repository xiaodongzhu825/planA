package _type

import "time"

// DelTaskDetail 删除任务详情表结构
type DelTaskDetail struct {
	ID         int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DelTaskID  *int64     `gorm:"column:del_task_id;default:0" json:"del_task_id"`       // 删除任务 id
	TaskID     *string    `gorm:"column:task_id;default:'';size:255" json:"task_id"`     // 任务 id
	Isbn       *string    `gorm:"column:isbn;default:'';size:255" json:"isbn"`           // isbn
	BookName   *string    `gorm:"column:book_name;default:'';size:255" json:"book_name"` // 图书名称
	Token      *string    `gorm:"column:token;default:'';size:255" json:"token"`         // token
	GoodsID    *int64     `gorm:"column:goods_id" json:"goods_id"`                       // 商品 id
	JSON       *string    `gorm:"column:json;type:text" json:"json"`                     // 原始字符串
	Status     *int64     `gorm:"column:status;default:0" json:"status"`                 // 状态: 1=正常 2=错误
	Err        *string    `gorm:"column:err;type:text" json:"err"`                       // 错误信息
	DeleteAt   *time.Time `gorm:"column:delete_at" json:"delete_at"`                     // 删除商品时间
	DeleteDate *string    `gorm:"column:delete_date" json:"delete_date"`                 // 删除商品日期
	CreateAt   *time.Time `gorm:"column:create_at" json:"create_at"`                     // 创建时间
}
