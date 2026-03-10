package service

import (
	"errors"
	"planA/initialization/golabl"
	"planA/tool"
	mysqlType "planA/type/mysql"

	"gorm.io/gorm"
)

// InsertTaskUser 向task_user表插入单条数据
// @param taskUser 要插入的task_user数据
// @return error 错误信息
func InsertTaskUser(taskUser *mysqlType.TaskUser) error {
	err := golabl.MysqlDb.Create(taskUser).Error
	return err
}

// PageQueryTaskUserByUserId 分页查询任务-用户关联表数据
// @param params 分页查询参数
// @return []*TaskUser 查询结果
// @return int64 总条数
// @return error 错误信息
func PageQueryTaskUserByUserId(params *mysqlType.PageQueryTaskUserByUserIdParams) ([]*mysqlType.TaskUser, int64, error) {
	var taskUsers []*mysqlType.TaskUser
	var total int64
	//获取页
	pageSize, offset := tool.GetPage(params.Page.PageNum, params.Page.PageSize)
	// 构建查询条件
	query := golabl.MysqlDb.Model(&mysqlType.TaskUser{}).Where("user_id = ?", params.UserID)
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
		return taskUsers, total, nil
	}
	// 分页查询数据
	err := query.Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&taskUsers).Error
	if err != nil {
		return nil, 0, err
	}
	return taskUsers, total, err
}

// GetTaskUserById 根据ID查询task_user表数据
// @param id 记录ID
// @return *TaskUser 查询结果
// @return error 错误信息
func GetTaskUserById(id int64) (mysqlType.TaskUser, error) {
	var taskUser mysqlType.TaskUser
	err := golabl.MysqlDb.First(&taskUser, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return taskUser, nil
	}
	if err != nil {
		return taskUser, err
	}
	return taskUser, nil
}

// GetTaskUserByTaskId 根据TaskId查询task_user表数据
// @param taskId 任务Id
// @return *TaskUser 查询结果
// @return error 错误信息
func GetTaskUserByTaskId(taskId string) (mysqlType.TaskUser, error) {
	var taskUser mysqlType.TaskUser

	err := golabl.MysqlDb.Where("task_id = ?", taskId).First(&taskUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return taskUser, nil
	}
	if err != nil {
		return taskUser, err
	}

	return taskUser, nil
}

// UpdateTaskUserIsExport 修改task_user表的is_export
// @param id 记录ID
// @param isExport 是否导出
// @return error 错误信息
func UpdateTaskUserIsExport(taskId string, isExport int) error {
	err := golabl.MysqlDb.Model(&mysqlType.TaskUser{}).Where("TASK_ID = ?", taskId).Update("is_export", isExport).Error
	return err
}
