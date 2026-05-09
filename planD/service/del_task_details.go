package service

import (
	"fmt"
	planBType "planA/planB/type"
	"planA/planD/initialization/golabl"
	"time"
)

// CreateTableIfNotExists 创建表
// @return error 错误信息
func CreateTableIfNotExists(taskId string) error {
	var dleTaskDetailsTable = fmt.Sprintf("del_task_details_%v", taskId)
	// 检查表是否存在
	if !golabl.MysqlDb.Migrator().HasTable(dleTaskDetailsTable) {
		sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
            id int(11) NOT NULL AUTO_INCREMENT,
            del_task_id int(11) DEFAULT '0' COMMENT '删除任务id',
            task_id varchar(255) COLLATE utf8_unicode_ci DEFAULT '' COMMENT '任务id',
            isbn varchar(255) COLLATE utf8_unicode_ci DEFAULT '' COMMENT 'isbn',
            book_name varchar(255) COLLATE utf8_unicode_ci DEFAULT '' COMMENT '商品名称',
            token varchar(255) COLLATE utf8_unicode_ci DEFAULT '' COMMENT 'token',
            goods_id bigint(11) DEFAULT NULL COMMENT '商品id',
            json text COLLATE utf8_unicode_ci COMMENT '原始字符串',
    		status int(11) DEFAULT '0' COMMENT '状态: 1=正常 2=错误',
    		err  text COLLATE utf8_unicode_ci COMMENT '错误信息',
            pdd_delete_at datetime DEFAULT NULL COMMENT '请求拼多多删除商品时间',
            pdd_delete_date date DEFAULT NULL COMMENT '请求拼多多删除商品日期',
            create_at datetime DEFAULT NULL COMMENT '创建时间',
            PRIMARY KEY (id),
            KEY del_task_id (del_task_id, task_id, goods_id)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci`, dleTaskDetailsTable)

		if err := golabl.MysqlDb.Exec(sql).Error; err != nil {
			return fmt.Errorf("创建 %v 表失败: %v", dleTaskDetailsTable, err)
		}
	}
	return nil
}

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

// InsertDelTaskDetail 插入单条删除任务详情数据
func InsertDelTaskDetail(delTaskID int64, taskId string, token string, bookName string, isbn string, goodsId int64) error {
	var dleTaskDetailsTable = fmt.Sprintf("del_task_details_%v", taskId)

	// 先检查并创建表
	if err := CreateTableIfNotExists(taskId); err != nil {
		return err
	}

	now := time.Now()
	status := int64(1)

	delTaskDetail := &planBType.DelTaskDetail{
		DelTaskID:     &delTaskID,
		TaskID:        &taskId,
		BookName:      &bookName,
		Token:         &token,
		Isbn:          &isbn,
		GoodsID:       &goodsId,
		Status:        &status,
		PddDeleteAt:   nil,
		PddDeleteDate: nil,
		CreateAt:      &now,
	}

	// 使用动态表名插入
	result := golabl.MysqlDb.Table(dleTaskDetailsTable).Create(delTaskDetail)
	if result.Error != nil {
		return fmt.Errorf("插入数据失败: %v", result.Error)
	}
	return nil
}
