package i

import (
	_type "planA/type"
)

type TaskExport interface {
	CreateTaskExport(export _type.TaskExportDTO) error                                          //创建导出任务
	GetTaskExportList(page, pageSize int, userId string) ([]*_type.TaskExportDTO, int64, error) //获取导出任务列表
	GetTaskExportByTaskId(taskId string) (_type.TaskExportDTO, error)                           //根据任务 ID获取导出任务
	UpdateTaskExport(export _type.TaskExportDTO) error                                          //更新导出任务
	UpdateTaskExportStatus(taskId string, status int64, fileUrl string) error
	GetTaskExportOldList() ([]*_type.TaskExportDTO, error) //获取导出任务旧数据
	DeleteTaskExportOldData() error                        //删除导出任务旧数据
}
