package i

import (
	_type "planA/type"
)

type TaskRecords interface {
	CreateTaskRecords(user _type.TaskRecordsDTO) error                                             //创建任务记录
	GetTaskRecordsList(params _type.GetTaskRecordsListReq) ([]*_type.TaskRecordsDTO, int64, error) //获取任务记录列表
	GetTaskRecordsByTaskId(taskId string) (*_type.TaskRecordsDTO, error)                           //根据任务 ID获取任务记录
	UpdateTaskRecords(user _type.TaskRecordsDTO) error                                             //更新任务记录
	GetTaskRecordsOldList() ([]_type.TaskRecordsDTO, error)                                        //获取任务记录旧数据列表
	DeleteTaskRecordsOldData() error                                                               //删除任务记录旧数据
	DeleteTaskRecordsByTaskId(taskId string) error                                                 //根据任务 ID删除任务记录
	GetTaskRecords24Hour() ([]*_type.TaskRecordsDTO, error)                                        //获取24小时内的数据
}
