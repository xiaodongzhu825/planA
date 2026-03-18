package mysql

import (
	"errors"
	"planA/initialization/golabl"
	"planA/tool"
	mysqlType "planA/type/mysql"
	"time"

	"gorm.io/gorm"
)

// CreateTaskRecords 向task_records表插入单条数据
// @param TaskRecords 要插入的task_records数据
// @return error 错误信息
func CreateTaskRecords(TaskRecords *mysqlType.TaskRecords) error {
	err := golabl.MysqlDb.Create(TaskRecords).Error
	return err
}

// GetTaskRecordsList 分页查询任务-用户关联表数据
// @param params 分页查询参数
// @return []*TaskRecords 查询结果
// @return int64 总条数
// @return error 错误信息
func GetTaskRecordsList(params *mysqlType.GetTaskRecordsByUserIdParams) ([]*mysqlType.TaskRecords, int64, error) {
	var TaskRecordss []*mysqlType.TaskRecords
	var total int64
	//获取页
	pageSize, offset := tool.GetPage(params.Page.PageNum, params.Page.PageSize)
	// 构建查询条件
	query := golabl.MysqlDb.Model(&mysqlType.TaskRecords{})
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.ShopName != "" {
		query = query.Where("shop_name = ?", params.ShopName)
	}
	if params.TaskID != "" {
		query = query.Where("task_id = ?", params.TaskID)
	}
	if params.TaskType != 0 {
		query = query.Where("task_type = ?", params.TaskType)
	}
	// 查询总条数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return TaskRecordss, total, nil
	}
	// 分页查询数据
	err := query.Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&TaskRecordss).Error
	if err != nil {
		return nil, 0, err
	}
	return TaskRecordss, total, err
}

// GetTaskRecordsByTaskId 根据TaskId查询task_records表数据
// @param taskId 任务Id
// @return *TaskRecords 查询结果
// @return error 错误信息
func GetTaskRecordsByTaskId(taskId string) (mysqlType.TaskRecords, error) {
	var TaskRecords mysqlType.TaskRecords

	err := golabl.MysqlDb.Where("task_id = ?", taskId).First(&TaskRecords).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return TaskRecords, nil
	}
	if err != nil {
		return TaskRecords, err
	}

	return TaskRecords, nil
}

// UpdateTaskRecords 根据任务ID更新数据
func UpdateTaskRecords(record *mysqlType.TaskRecords) error {
	return golabl.MysqlDb.Model(&mysqlType.TaskRecords{}).Where("task_id = ?", record.TaskID).Updates(record).Error
}

// DeleteOldTaskRecords 删除大于三天的数据
// @return error 错误信息
func DeleteOldTaskRecords() error {
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	return golabl.MysqlDb.Where("create_at < ?", threeDaysAgo).Delete(&mysqlType.TaskRecords{}).Error
}

// DeleteTaskRecordsByTaskId 根据任务ID删除数据
// @param taskId 任务ID
// @return error 错误信息
func DeleteTaskRecordsByTaskId(taskId string) error {
	return golabl.MysqlDb.Where("task_id = ?", taskId).Delete(&mysqlType.TaskRecords{}).Error
}

// GetTaskRecords24Hour 获取24小时内的数据
func GetTaskRecords24Hour() ([]*mysqlType.TaskRecords, error) {
	var tasks []*mysqlType.TaskRecords
	now := time.Now()
	twentyFourHoursAgo := now.Add(-24 * time.Hour)
	tenMinutesAgo := now.Add(-10 * time.Minute)

	err := golabl.MysqlDb.Where("create_at >= ? AND create_at <= ?",
		twentyFourHoursAgo, tenMinutesAgo).
		Order("create_at DESC").
		Find(&tasks).Error

	return tasks, err
}
