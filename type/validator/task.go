package validator

// CreateTask 创建任务结构体
type CreateTask struct {
	ShopID    string `form:"shop_id" validate:"required,min=3,max=20"`     // 必填，长度3-20
	ShopType  string `form:"shop_type" validate:"required,oneof=1 2 5"`    // 必填，只能是1、2、5
	TaskCount string `form:"task_count" validate:"required,numeric,min=1"` // 必填，数字且最小值为1
	TaskType  string `form:"task_type" validate:"required,oneof=1 2"`      // 必填，只能是1或2
	ImgType   string `form:"img_type" validate:"required,oneof=1 2 3 4"`   // 必填，只能是1、2、3、4
}

// UpdateTaskStatus 更改任务状态结构体
type UpdateTaskStatus struct {
	TaskID string `form:"task_id" validate:"required"` //必填
}

// GetTask 获取任务列表结构体
type GetTask struct {
	Page     string `form:"page"`
	Size     string `form:"size"`
	TaskID   string `form:"task_id"`
	ShopName string `form:"shop_name"`
	TaskType string `form:"task_type"`
}

// GetTaskByUserId 获取用户任务列表结构体
type GetTaskByUserId struct {
	Page     string `form:"page"`
	Size     string `form:"size"`
	TaskID   string `form:"task_id"`
	ShopName string `form:"shop_name"`
	TaskType string `form:"task_type"`
	UserID   string `form:"user_id" validate:"required"` //必填
}
