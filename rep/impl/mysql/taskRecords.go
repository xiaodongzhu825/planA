package mysql

import (
	mysqlServer "planA/service/mysql"
	_type "planA/type"
	mysqlType "planA/type/mysql"
	"time"
)

// CreateTaskRecords 创建任务记录
// @param export 任务记录
// @return error 错误信息
func (g *GormAdapter) CreateTaskRecords(export _type.TaskRecordsDTO) error {
	return mysqlServer.CreateTaskRecords(&mysqlType.TaskRecords{
		UserID:   &export.UserId,
		ShopID:   &export.ShopId,
		TaskID:   &export.TaskId,
		ShopName: &export.ShopName,
		IsExport: &export.IsExport,
		TaskType: &export.TaskType,
	})
}

// GetTaskRecordsList 获取任务记录列表
// @param params 查询参数
// @return []mysqlType.TaskExport 任务记录列表
// @return error 错误信息
func (g *GormAdapter) GetTaskRecordsList(params _type.GetTaskRecordsListReq) ([]*_type.TaskRecordsDTO, int64, error) {
	list, count, err := mysqlServer.GetTaskRecordsList(&mysqlType.GetTaskRecordsByUserIdParams{
		UserID:   params.UserId,
		ShopName: params.ShopName,
		TaskID:   params.TaskId,
		TaskType: params.TaskType,
		Page: _type.Page{
			PageNum:  params.Page,
			PageSize: params.Size,
		},
	})
	listDTO := convertMysqlTaskRecordsToDTO(list)
	return listDTO, count, err
}

// GetTaskRecordsByTaskId 根据任务ID获取任务记录
// @param taskId 任务ID
// @return *mysqlType.TaskExport 任务记录
// @return error 错误信息
func (g *GormAdapter) GetTaskRecordsByTaskId(taskId string) (*_type.TaskRecordsDTO, error) {
	info, err := mysqlServer.GetTaskRecordsByTaskId(taskId)
	infoDTO := _type.TaskRecordsDTO{
		Id:       info.ID,
		UserId:   *info.UserID,
		ShopId:   *info.ShopID,
		TaskId:   *info.TaskID,
		ShopName: *info.ShopName,
		IsExport: *info.IsExport,
		TaskType: *info.TaskType,
		CreateAt: time.Now(),
	}
	return &infoDTO, err
}

// UpdateTaskRecords 更新任务记录
// @param export 任务记录
// @return error 错误信息
func (g *GormAdapter) UpdateTaskRecords(user _type.TaskRecordsDTO) error {
	return mysqlServer.UpdateTaskRecords(&mysqlType.TaskRecords{
		UserID:   &user.UserId,
		ShopID:   &user.ShopId,
		TaskID:   &user.TaskId,
		ShopName: &user.ShopName,
		IsExport: &user.IsExport,
		TaskType: &user.TaskType,
	})
}

// GetTaskRecordsOldList 获取任务记录旧数据列表
// @return *mysqlType.TaskExport 任务记录列表
// @return error 错误信息
func (g *GormAdapter) GetTaskRecordsOldList() ([]_type.TaskRecordsDTO, error) {
	return nil, nil
}

// DeleteTaskRecordsOldData 删除任务记录旧数据
// @param taskId 任务ID
// @return error 错误信息
func (g *GormAdapter) DeleteTaskRecordsOldData() error {
	return mysqlServer.DeleteOldTaskRecords()
}

// DeleteTaskRecordsByTaskId 根据任务 ID删除任务记录
func (g *GormAdapter) DeleteTaskRecordsByTaskId(taskId string) error {
	return mysqlServer.DeleteTaskRecordsByTaskId(taskId)
}

// GetTaskRecords24Hour 获取24小时内的任务记录
func (g *GormAdapter) GetTaskRecords24Hour() ([]*_type.TaskRecordsDTO, error) {
	list, err := mysqlServer.GetTaskRecords24Hour()
	listDTO := convertMysqlTaskRecordsToDTO(list)
	return listDTO, err
}

// convertMysqlToDTO 转换mysqlType.TaskExport为_type.TaskExport
// @param records mysqlType.TaskExport列表
// @return []*_type.TaskExport _type.TaskExport列表
// @return error 错误信息
func convertMysqlTaskRecordsToDTO(records []*mysqlType.TaskRecords) []*_type.TaskRecordsDTO {
	dtos := make([]*_type.TaskRecordsDTO, len(records))
	for i, r := range records {
		dtos[i] = &_type.TaskRecordsDTO{
			Id:       r.ID,
			UserId:   *r.UserID,
			ShopId:   *r.ShopID,
			TaskId:   *r.TaskID,
			ShopName: *r.ShopName,
			IsExport: *r.IsExport,
			TaskType: *r.TaskType,
			CreateAt: *r.CreateAt,
		}
	}
	return dtos
}
