package mysql

import (
	"fmt"
	"planA/initialization/golabl"
	planBType "planA/planB/type"
	mysqlType "planA/type/mysql"
	"time"

	"gorm.io/gorm"
)

// GetDelTask 查询del_task 表中待执行
func GetDelTask() ([]mysqlType.DelTask, error) {
	var delTask []mysqlType.DelTask
	err := golabl.MysqlDb.Where("status IN ? and shop_type != 6", []int{0, 1, 2}).Find(&delTask).Error
	if err != nil {
		fmt.Println("查询del_task 表中待执行失败：", err)
	}
	return delTask, err
}

// UpdateDelTaskStatus 根据id 将查询del_task 的status 修改为 1
func UpdateDelTaskStatus(id int64) error {
	err := golabl.MysqlDb.Model(&mysqlType.DelTask{}).Where("id = ?", id).Update("status", 1).Error
	return err
}

// GetDelTaskByTaskId 根据tasId查询删除任务
func GetDelTaskByTaskId(taskId string) (mysqlType.DelTask, error) {
	var delTask mysqlType.DelTask
	err := golabl.MysqlDb.Where("task_id = ?", taskId).First(&delTask).Error
	return delTask, err
}

// UpdateDelTaskPidByTaskId 根据tasId 清空 pid
func UpdateDelTaskPidByTaskId(taskId string) error {
	err := golabl.MysqlDb.Model(&mysqlType.DelTask{}).Where("task_id = ?", taskId).Update("pid", "").Error
	return err
}

// UpdateDelTaskPidByTaskIdAndPid 根据tasId 设置 pid
func UpdateDelTaskPidByTaskIdAndPid(taskId string, pid string) error {
	err := golabl.MysqlDb.Model(&mysqlType.DelTask{}).Where("task_id = ?", taskId).Update("pid", pid).Error
	return err
}

// GetDelTaskByPage 分页查询
func GetDelTaskByPage(pageNum int, pageSize int, userId string) ([]mysqlType.DelTask, int, error) {
	var delTask []mysqlType.DelTask
	var total int64

	// 构建基础查询
	query := golabl.MysqlDb.Model(&mysqlType.DelTask{})

	// 如果userId不为空，添加where条件
	if userId != "" {
		query = query.Where("user_id = ?", userId)
	}

	// 查询总记录数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 如果总数为0，直接返回空切片
	if total == 0 {
		return []mysqlType.DelTask{}, 0, nil
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 检查偏移量是否超出范围
	if offset >= int(total) {
		return []mysqlType.DelTask{}, int(total), nil
	}

	// 分页查询数据
	err := query.
		Limit(pageSize).
		Offset(offset).
		Order("id DESC"). // 建议添加排序，保证分页顺序稳定
		Find(&delTask).Error

	if err != nil {
		return nil, 0, fmt.Errorf("分页查询失败: %w", err)
	}

	return delTask, int(total), nil
}

// GetDelTaskDetailByPage 查询删除任务详情
func GetDelTaskDetailByPage(pageNum int, pageSize int, taskId string, status int) ([]planBType.DelTaskDetail, int, error) {
	var delTaskDetail []planBType.DelTaskDetail
	var total int64
	tableName := "del_task_details_" + taskId
	// 构建基础查询
	query := golabl.MysqlDb.Table(tableName)

	if status != 3 {
		query = query.Where("status = ?", status)
	}

	// 查询总记录数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 如果总数为0，直接返回空切片
	if total == 0 {
		return []planBType.DelTaskDetail{}, 0, nil
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 检查偏移量是否超出范围
	if offset >= int(total) {
		return []planBType.DelTaskDetail{}, int(total), nil
	}

	// 分页查询数据
	err := query.
		Limit(pageSize).
		Offset(offset).
		Order("id DESC"). // 建议添加排序，保证分页顺序稳定
		Find(&delTaskDetail).Error

	if err != nil {
		return nil, 0, fmt.Errorf("分页查询失败: %w", err)
	}

	return delTaskDetail, int(total), nil
}

// GetExpiredDelTask 查询过期的删除任务
func GetExpiredDelTask() ([]mysqlType.DelTask, error) {
	var delTask []mysqlType.DelTask
	// 假设过期时间是7天前，根据实际需求调整
	expireTime := time.Now().Add(-7 * 24 * time.Hour)
	err := golabl.MysqlDb.Where("status = 3 and create_at < ?", expireTime).Find(&delTask).Error
	return delTask, err
}

// DeleteDelTaskDetail 删除任务明细表
func DeleteDelTaskDetail(taskId string) error {
	tableName := "del_task_details_" + taskId
	err := golabl.MysqlDb.Migrator().DropTable(tableName)
	return err
}

// DeleteDelTaskById 根据id 删除 删除任务中的数据
func DeleteDelTaskById(id int64) error {
	err := golabl.MysqlDb.Where("id = ?", id).Delete(&mysqlType.DelTask{}).Error
	return err
}

// CreateDelTask 创建task 删除任务
func CreateDelTask(data mysqlType.DelTask) error {
	return golabl.MysqlDb.Create(&data).Error
}

// UpdateDelTaskProgress 修改任务进度
func UpdateDelTaskProgress(taskId string, num int) error {
	// 使用 SQL 表达式在原有值基础上累加
	return golabl.MysqlDb.Model(&mysqlType.DelTask{}).
		Where("task_id = ?", taskId).
		Update("task_count_over", gorm.Expr("task_count_over + ?", num)).Error
}

// UpdateDelTaskStatusByTaskId 修改任务状态
func UpdateDelTaskStatusByTaskId(taskId string, status int) error {
	return golabl.MysqlDb.Model(&mysqlType.DelTask{}).Where("task_id = ?", taskId).Update("status", status).Error
}
