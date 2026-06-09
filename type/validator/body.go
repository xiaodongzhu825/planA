package validator

// InsterBodyOver 像bodyOver中插入一条数据
type InsterBodyOver struct {
	TaskId string `form:"task_id" validate:"required"` //必填
	Data   string `form:"data" validate:"required"`    //必填
}

// GetTaskIdToBody 获取与body
type GetTaskIdToBody struct {
	TaskId string `form:"task_id" validate:"required"` //必填
}
