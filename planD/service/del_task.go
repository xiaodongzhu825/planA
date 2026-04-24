package service

import (
	"planA/planD/initialization/golabl"
	planAType "planA/type/mysql"
	"time"

	"gorm.io/gorm"
)

// GetDelTask 查询指定的删除任务
func GetDelTask() (planAType.DelTask, error) {
	var delTask planAType.DelTask
	err := golabl.MysqlDb.Where("task_id = ?", golabl.TaskId).First(&delTask).Error
	return delTask, err
}

// UpdateTaskCountOver 根据task_id 将 task_count_over +1
func UpdateTaskCountOver() error {
	return golabl.MysqlDb.Model(&planAType.DelTask{}).
		Where("task_id = ?", golabl.TaskId).
		Updates(map[string]interface{}{
			"task_count_over": gorm.Expr("task_count_over + 1"),
		}).Error
}

// UpdateTaskStatus 根据task_id 修改 status
func UpdateTaskStatus(status int) error {
	updateData := map[string]interface{}{
		"status": status,
	}

	// 如果 status=2，设置 pause_at 为当前时间
	if status == 2 {
		updateData["pause_at"] = time.Now()
	}

	return golabl.MysqlDb.Model(&planAType.DelTask{}).
		Where("task_id = ?", golabl.TaskId).
		Updates(updateData).Error
}
