package service

import (
	"errors"
	"planA/planB/initialization/golabl"
	planATypeMysql "planA/type/mysql"

	"gorm.io/gorm"
)

// GetDelTaskByTaskId 获取当前删除任务
// @return planATypeMysql.DelTask 删除任务信息
// @return bool 是否存在
// @return error 错误信息
func GetDelTaskByTaskId() (planATypeMysql.DelTask, bool, error) {
	var delTask planATypeMysql.DelTask
	err := golabl.MysqlDb.Where("task_id = ?", golabl.Task.TaskId).First(&delTask).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在
			return delTask, false, nil
		}
		// 其他错误
		return delTask, false, err
	}

	// 记录存在
	return delTask, true, nil
}

// CreateDelTask 创建删除任务
// @param delTask 删除任务信息
// @return planATypeMysql.DelTask 删除任务信息
// @return error 错误信息
func CreateDelTask(delTask planATypeMysql.DelTask) (planATypeMysql.DelTask, error) {
	err := golabl.MysqlDb.Create(&delTask).Error
	return delTask, err
}

// AddDelTaskDetailCount 增加删除任务详情的任务数
// @return error 错误信息
func AddDelTaskDetailCount() error {
	return golabl.MysqlDb.Model(&planATypeMysql.DelTask{}).Where("task_id = ?", golabl.Task.TaskId).Update("task_count", gorm.Expr("task_count + ?", 1)).Error
}

// UpdateDelTaskStatusToDoing 将完成的任务状态修改为执行中
func UpdateDelTaskStatusToDoing() error {
	return golabl.MysqlDb.Model(&planATypeMysql.DelTask{}).Where("task_id = ?", golabl.Task.TaskId).Update("status", 1).Error
}
