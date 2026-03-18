package validator

// GetExportTask 获取导出任务列表结构体
type GetExportTask struct {
	Page string `form:"page"`
	Size string `form:"size"`
}

// GetExportTaskByUserId 获取导出任务列表结构体-用户
type GetExportTaskByUserId struct {
	UserID string `form:"user_id" validate:"required"`
	Page   string `form:"page"`
	Size   string `form:"size"`
}

// ExportTaskDetail 根据任务 id导出任务详情结构体
type ExportTaskDetail struct {
	TaskID string `form:"task_id" validate:"required"` //必填
}

// ExportTaskDetailByUserId 根据任务 id导出任务详情结构体-用户
type ExportTaskDetailByUserId struct {
	TaskID string `form:"task_id" validate:"required"` //必填
	UserID string `form:"user_id" validate:"required"` //必填
}
