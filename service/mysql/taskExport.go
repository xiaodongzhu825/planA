package mysql

import (
	"database/sql"
	"planA/initialization/golabl"
	"planA/tool"
	mysqlType "planA/type/mysql"
	"time"
)

// CreateTaskExport 向task_export表插入一条记录
// @param export TaskExport 要插入的导出记录
// @return int64 插入记录的自增ID
// @return error 错误信息

func CreateTaskExport(export mysqlType.TaskExport) (int64, error) {
	// 创建记录
	createAt := time.Now()
	export.CreateAt = &createAt
	result := golabl.MysqlDb.Model(&mysqlType.TaskExport{}).Create(&export)
	// 检查是否有错误
	if result.Error != nil {
		return 0, result.Error
	}
	// 返回插入的自增 ID
	return export.ID, nil
}

// UpdateTaskExportStatus 更新task_export表中的status字段
// @param taskId string 任务ID
// @param status int64 状态
// @param fullPath string 文件路径
// @return error 错误信息

func UpdateTaskExportStatus(taskId string, status int64, fullPath string) error {
	var err error
	if status == 2 {
		completeAt := &sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
		err = golabl.MysqlDb.Model(&mysqlType.TaskExport{}).Where("task_id = ?", taskId).Updates(&mysqlType.TaskExport{
			FileUrl:    &fullPath,
			Status:     &status,
			CompleteAt: completeAt,
		}).Error
	} else {
		err = golabl.MysqlDb.Model(&mysqlType.TaskExport{}).Where("task_id = ?", taskId).Update("status", status).Error
	}
	return err
}

// GetTaskExportList 分页查询task_export表记录
// @param page 分页参数
// @param pageSize int 每页数量
// @param userId string 用户ID
// @return []*TaskUser 查询结果
// @return int64 总条数
// @return error 错误信息
func GetTaskExportList(page, pageSize int, userId string) ([]mysqlType.TaskExport, int64, error) {
	var taskExport []mysqlType.TaskExport
	var total int64
	//获取页
	pageSize, offset := tool.GetPage(page, pageSize)
	// 构建查询条件
	query := golabl.MysqlDb.Model(&mysqlType.TaskExport{})
	if userId != "" {
		query = query.Where("user_id = ?", userId)
	}
	// 查询总条数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return taskExport, total, nil
	}
	// 分页查询数据
	err := query.Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&taskExport).Error
	if err != nil {
		return nil, 0, err
	}
	return taskExport, total, err
}

// DeleteOldExport 删除task_export表中3天前的记录
// @return error 错误信息
func DeleteOldExport() error {
	// 计算三天前的日期时间
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	return golabl.MysqlDb.Where("create_at < ?", threeDaysAgo).Delete(&mysqlType.TaskExport{}).Error
}

// GetOldExportSQLite 获取task_export表中3天前的记录
func GetOldExportSQLite() ([]mysqlType.TaskExport, error) {
	var taskExport []mysqlType.TaskExport
	// 计算三天前的日期时间
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	err := golabl.MysqlDb.Where("create_at < ?", threeDaysAgo).Find(&taskExport).Error
	return taskExport, err
}
