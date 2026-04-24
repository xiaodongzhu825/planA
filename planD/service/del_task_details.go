package service

import (
	planBType "planA/planB/type"
	"planA/planD/initialization/golabl"
	"time"
)

// GetMax5000WaitDelTask 查询最大5000条等待删除的任务
func GetMax5000WaitDelTask() ([]*planBType.DelTaskDetail, error) {
	var delTask []*planBType.DelTaskDetail
	err := golabl.MysqlDb.Table("del_task_details_"+golabl.TaskId).Where("status = ?", 0).Limit(5000).Find(&delTask).Error
	return delTask, err
}

// UpdateDelTaskDetailStatus 根据goodsId修改数据状态
func UpdateDelTaskDetailStatus(id int, status int, err string) error {
	now := time.Now() // 获取当前时间
	return golabl.MysqlDb.Table("del_task_details_"+golabl.TaskId).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          status,
			"err":             err,
			"pdd_delete_at":   now,
			"pdd_delete_date": now.Format("2006-01-02"),
		}).Error
}

// GetTaskCountOver 获取未完成的任务数量
func GetTaskCountOver() (int64, error) {
	var count int64
	err := golabl.MysqlDb.Table("del_task_details_"+golabl.TaskId).Where("status = ?", 0).Count(&count).Error
	return count, err
}
