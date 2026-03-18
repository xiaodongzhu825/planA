package mysql

import (
	"database/sql"
	mysqlServer "planA/service/mysql"
	_type "planA/type"
	mysqlType "planA/type/mysql"
)

// CreateTaskExport 创建任务导出
// @param export 任务导出
// @return error 错误信息
func (g *GormAdapter) CreateTaskExport(export _type.TaskExportDTO) error {
	_, err := mysqlServer.CreateTaskExport(mysqlType.TaskExport{
		ID:         export.Id,
		UserID:     &export.UserId,
		ShopID:     &export.ShopId,
		TaskID:     &export.TaskId,
		ShopName:   &export.ShopName,
		FileUrl:    &export.FileUrl,
		Status:     &export.Status,
		Total:      &export.Total,
		CompleteAt: &export.CompleteAt,
	})
	return err
}

// GetTaskExportList 获取任务导出列表
// @param params 查询参数
// @return []mysqlType.TaskExportDTO 任务导出列表
// @return error 错误信息
func (g *GormAdapter) GetTaskExportList(page, pageSize int, userId string) ([]*_type.TaskExportDTO, int64, error) {
	list, count, err := mysqlServer.GetTaskExportList(page, pageSize, userId)
	listDTO := convertMysqlTaskExportToDTO(list)
	return listDTO, count, err
}

// GetTaskExportByTaskId 根据任务 ID获取导出任务
// @param taskId 任务 ID
// @return *mysqlType.TaskExportDTO 导出任务
// @return error 错误信息
func (g *GormAdapter) GetTaskExportByTaskId(taskId string) (_type.TaskExportDTO, error) {
	return _type.TaskExportDTO{}, nil
}

// UpdateTaskExport 更新导出任务
// @param export 导出任务
// @return error 错误信息
func (g *GormAdapter) UpdateTaskExport(export _type.TaskExportDTO) error {
	return nil
}

// GetTaskExportOldList 获取任务导出旧数据列表
func (g *GormAdapter) GetTaskExportOldList() ([]*_type.TaskExportDTO, error) {
	list, err := mysqlServer.GetOldExportSQLite()
	listDTO := convertMysqlTaskExportToDTO(list)
	return listDTO, err
}

// DeleteTaskExportOldData 删除任务导出旧数据
func (g *GormAdapter) DeleteTaskExportOldData() error {
	return mysqlServer.DeleteOldExport()
}

// UpdateTaskExportStatus 更新任务导出状态
// @param taskId 任务 ID
// @param status 状态
// @param fileUrl 文件路径
// @return error 错误信息
func (s *GormAdapter) UpdateTaskExportStatus(taskId string, status int64, fileUrl string) error {
	return mysqlServer.UpdateTaskExportStatus(taskId, status, fileUrl)
}

func convertMysqlTaskExportToDTO(records []mysqlType.TaskExport) []*_type.TaskExportDTO {
	dtos := make([]*_type.TaskExportDTO, len(records))
	for i, r := range records {
		dtos[i] = &_type.TaskExportDTO{
			Id:       r.ID,
			UserId:   *r.UserID,
			ShopId:   *r.ShopID,
			TaskId:   *r.TaskID,
			ShopName: *r.ShopName,
			FileUrl:  *r.FileUrl,
			Status:   *r.Status,
			Total:    *r.Total,
			// 安全处理 CompleteAt
			CompleteAt: func() sql.NullTime {
				if r.CompleteAt != nil {
					return *r.CompleteAt
				}
				return sql.NullTime{Valid: false}
			}(),
			CreateAt: *r.CreateAt,
		}
	}
	return dtos
}
