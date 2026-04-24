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
