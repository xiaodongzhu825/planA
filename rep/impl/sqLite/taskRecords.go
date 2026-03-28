package sqLite

import (
	sqLiteServer "planA/service/sqLite"
	_type "planA/type"
	sqliteType "planA/type/sqLite"
	"time"
)

// CreateTaskRecords 创建任务记录
// @param export 任务记录
// @return error 错误信息
func (s *SqlAdapter) CreateTaskRecords(records _type.TaskRecordsDTO) error {
	return sqLiteServer.CreateTaskRecords(sqliteType.TaskRecords{
		UserID:   records.UserId,
		ShopID:   records.ShopId,
		TaskID:   records.TaskId,
		ShopName: records.ShopName,
		IsExport: records.IsExport,
		TaskType: records.TaskType,
		CreateAt: time.Time{},
	})
}

// GetTaskRecordsList 获取任务记录列表
// @param params 查询参数
// @return []mysqlType.TaskExport 任务记录列表
// @return error 错误信息
func (s *SqlAdapter) GetTaskRecordsList(params _type.GetTaskRecordsListReq) ([]*_type.TaskRecordsDTO, int64, error) {
	list, count, err := sqLiteServer.GetTaskRecordsList(sqliteType.GetTaskRecordsByUserIdParams{
		UserID:   params.UserId,
		ShopName: params.ShopName,
		TaskID:   params.TaskId,
		TaskType: params.TaskType,
		Page: _type.Page{
			PageNum:  params.Page,
			PageSize: params.Size,
		},
	})
	listDTO := convertSqliteTaskRecordsToDTO(list)
	return listDTO, count, err
}

// GetTaskRecordsByTaskId 根据任务ID获取任务记录
func (s *SqlAdapter) GetTaskRecordsByTaskId(taskId string) (*_type.TaskRecordsDTO, error) {
	info, err := sqLiteServer.GetTaskRecordByTaskID(taskId)
	infoDTO := _type.TaskRecordsDTO{
		Id:       info.ID,
		UserId:   info.UserID,
		ShopId:   info.ShopID,
		TaskId:   info.TaskID,
		ShopName: info.ShopName,
		IsExport: info.IsExport,
		TaskType: info.TaskType,
		CreateAt: time.Now(),
	}
	return &infoDTO, err
}

// UpdateTaskRecords 更新任务记录
// @param export 任务记录
// @return error 错误信息
func (s *SqlAdapter) UpdateTaskRecords(user _type.TaskRecordsDTO) error {
	return sqLiteServer.UpdateTaskRecord(sqliteType.TaskRecords{
		ID:       user.Id,
		UserID:   user.UserId,
		ShopID:   user.ShopId,
		TaskID:   user.TaskId,
		ShopName: user.ShopName,
		IsExport: user.IsExport,
		TaskType: user.TaskType,
	})
}

// GetTaskRecordsOldList 获取任务记录旧数据列表
// @return *mysqlType.TaskExport 任务记录列表
// @return error 错误信息
func (s *SqlAdapter) GetTaskRecordsOldList() ([]*_type.TaskRecordsDTO, error) {
	list, err := sqLiteServer.GetTaskRecordsOldList()
	listDTO := convertSqliteTaskRecordsToDTO(list)
	return listDTO, err
}

// DeleteTaskRecordsOldData 删除任务记录旧数据
// @return error 错误信息
func (s *SqlAdapter) DeleteTaskRecordsOldData() error {

	return sqLiteServer.DeleteOldTaskRecords()
}

// DeleteTaskRecordsByTaskId 根据任务 ID删除任务记录
func (s *SqlAdapter) DeleteTaskRecordsByTaskId(taskId string) error {
	return sqLiteServer.DeleteTaskRecordsByTaskID(taskId)
}

// GetTaskRecords24Hour 获取24小时内的任务记录
func (s *SqlAdapter) GetTaskRecords24Hour() ([]*_type.TaskRecordsDTO, error) {
	list, err := sqLiteServer.GetTaskRecords24Hour()
	listDTO := convertSqliteTaskRecordsToDTO(list)
	return listDTO, err
}
func convertSqliteTaskRecordsToDTO(records []sqliteType.TaskRecords) []*_type.TaskRecordsDTO {
	dtos := make([]*_type.TaskRecordsDTO, len(records))
	for i, r := range records {
		dtos[i] = &_type.TaskRecordsDTO{
			Id:       r.ID,
			UserId:   r.UserID,
			ShopId:   r.UserID,
			TaskId:   r.TaskID,
			ShopName: r.ShopName,
			IsExport: r.IsExport,
			TaskType: r.TaskType,
			CreateAt: r.CreateAt,
		}
	}
	return dtos
}
