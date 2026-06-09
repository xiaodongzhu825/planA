package validator

// GetDelTask 获取任务列表结构体
type GetDelTask struct {
	Page string `form:"page"`
	Size string `form:"size"`
}

// GetDelTaskByUserId 获取任务列表结构体
type GetDelTaskByUserId struct {
	UserId string `form:"userId"`
	Page   string `form:"page"`
	Size   string `form:"size"`
}

// GetDelTaskDetail 获取任务详情列表
type GetDelTaskDetail struct {
	TaskId string `form:"taskId"`
	Page   string `form:"page"`
	Size   string `form:"size"`
}

// CreateTbDelTask 创建淘宝删除任务结构体
type CreateTbDelTask struct {
	ShopID   string `form:"shop_id" validate:"required,min=3,max=20"`
	TaskType string `form:"task_type" validate:"required,oneof=1 2 3"`
}

// CreateTbDelTaskDetails 插入淘宝删除任务数据c
type CreateTbDelTaskDetails struct {
	TaskID   string `form:"task_id" validate:"required"`
	Isbn     string `form:"isbn" validate:"required"`
	BookName string `form:"title" validate:"required"`
	GoodsId  string `form:"goods_id" validate:"required"`
	Status   string `form:"status"`
	Err      string `form:"err" validate:"required"`
}

// UpdateTbDelTaskDetailsStatus 修改指定淘宝删除任务详情状态
type UpdateTbDelTaskDetailsStatus struct {
	TaskID  string `form:"task_id" validate:"required"`
	GoodsId string `form:"goods_id" validate:"required"`
	Status  string `form:"status" validate:"required,oneof=1 2"`
	Err     string `form:"err" validate:"required"`
}

// UpdateTbDelTaskProgress 修改删除任务进度验证
type UpdateTbDelTaskProgress struct {
	TaskID string `form:"task_id" validate:"required"`
	Num    string `form:"num" validate:"required"`
}

// UpdateTbDelTaskStatus 修改删除任务状态验证
type UpdateTbDelTaskStatus struct {
	TaskID string `form:"task_id" validate:"required"`
	Status string `form:"status" validate:"required,oneof=1 2"`
}
