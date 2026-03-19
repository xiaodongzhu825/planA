package sqLite

import (
	sqLiteServer "planA/service/sqLite"
	_type "planA/type"
	sqliteType "planA/type/sqLite"
)

// CreateTaskExport 创建任务导出表
// @param export 任务导出表
// @return error 错误信息
func (s *SqlAdapter) CreateTaskExport(export _type.TaskExportDTO) error {
	_, err := sqLiteServer.CreateTaskExport(sqliteType.TaskExport{
		UserID:     export.UserId,
		ShopID:     export.ShopId,
		TaskID:     export.TaskId,
		ShopName:   export.ShopName,
		FileUrl:    export.FileUrl,
		Status:     export.Status,
		Total:      export.Total,
		CompleteAt: export.CompleteAt,
	})
	return err
}

// GetTaskExportList 获取任务导出列表
// @param page 分页
// @param pageSize 每页数量
// @param userId 用户ID
// @return []mysqlType.TaskExportDTO 任务导出列表
// @return error 错误信息
func (s *SqlAdapter) GetTaskExportList(page, pageSize int, userId string) ([]*_type.TaskExportDTO, int64, error) {
	list, count, err := sqLiteServer.GetTaskExportsList(page, pageSize, userId)
	listDTO := convertSqliteTaskExportToDTO(list)
	return listDTO, count, err
}

// GetTaskExportByTaskId 根据任务 ID获取导出任务
// @param taskId 任务 ID
// @return *mysqlType.TaskExportDTO 导出任务
// @return error 错误信息
func (s *SqlAdapter) GetTaskExportByTaskId(taskId string) (_type.TaskExportDTO, error) {
	info, err := sqLiteServer.GetTaskExportByTaskID(taskId)
	infoDTO := _type.TaskExportDTO{
		Id:         info.ID,
		UserId:     info.UserID,
		ShopId:     info.ShopID,
		TaskId:     info.TaskID,
		ShopName:   info.ShopName,
		FileUrl:    info.FileUrl,
		Status:     info.Status,
		Total:      info.Total,
		CompleteAt: info.CompleteAt,
		CreateAt:   info.CreateAt,
	}
	return infoDTO, err
}

// UpdateTaskExport 更新导出任务
// @param export 导出任务
// @return error 错误信息
func (s *SqlAdapter) UpdateTaskExport(export _type.TaskExportDTO) error {
	err := sqLiteServer.UpdateTaskExport(sqliteType.TaskExport{
		UserID:     export.UserId,
		ShopID:     export.ShopId,
		TaskID:     export.TaskId,
		ShopName:   export.ShopName,
		FileUrl:    export.FileUrl,
		Status:     export.Status,
		Total:      export.Total,
		CompleteAt: export.CompleteAt,
	})
	return err
}

// GetTaskExportOldList 获取任务导出旧数据列表
func (s *SqlAdapter) GetTaskExportOldList() ([]*_type.TaskExportDTO, error) {
	list, err := sqLiteServer.GetOldExport()
	listDTO := convertSqliteTaskExportToDTO(list)
	return listDTO, err
}

// DeleteTaskExportOldData 删除任务导出旧数据
func (s *SqlAdapter) DeleteTaskExportOldData() error {
	return sqLiteServer.DeleteOldExport()
}

// UpdateTaskExportStatus 更新任务导出状态
// @param taskId 任务 ID
// @param status 状态
// @param fileUrl 文件路径
// @return error 错误信息
func (s *SqlAdapter) UpdateTaskExportStatus(taskId string, status int64, fileUrl string) error {
	return sqLiteServer.UpdateTaskExportStatus(taskId, status, fileUrl)
}

func convertSqliteTaskExportToDTO(records []sqliteType.TaskExport) []*_type.TaskExportDTO {
	dtos := make([]*_type.TaskExportDTO, len(records))
	for i, r := range records {
		dtos[i] = &_type.TaskExportDTO{
			Id:         r.ID,
			UserId:     r.UserID,
			ShopId:     r.ShopID,
			TaskId:     r.TaskID,
			ShopName:   r.ShopName,
			FileUrl:    r.FileUrl,
			Status:     r.Status,
			Total:      r.Total,
			CompleteAt: r.CompleteAt,
			CreateAt:   r.CreateAt,
		}
	}
	return dtos
}
