package service

import (
	"encoding/json"
	"fmt"
	"planA/planB/initialization/golabl"
	planBType "planA/planB/type"
	planAType "planA/type"
	"time"
)

// CreateTableIfNotExists 创建表
// @return error 错误信息
func CreateTableIfNotExists() error {
	var dleTaskDetailsTable = fmt.Sprintf("del_task_details_%v", golabl.Task.TaskId)
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
            delete_at datetime DEFAULT NULL COMMENT '请求删除商品时间',
            delete_date date DEFAULT NULL COMMENT '请求删除商品日期',
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

// InsertDelTaskDetail 插入单条删除任务详情数据
func InsertDelTaskDetail(delTaskID int64, detail planAType.TaskBody) error {
	var dleTaskDetailsTable = fmt.Sprintf("del_task_details_%v", golabl.Task.TaskId)

	// 先检查并创建表
	if err := CreateTableIfNotExists(); err != nil {
		return err
	}

	now := time.Now()

	//将detail转为 json
	jsonByte, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	jsonStr := string(jsonByte)

	delTaskDetail := &planBType.DelTaskDetail{
		DelTaskID:  &delTaskID,
		TaskID:     &golabl.Task.TaskId,
		BookName:   &detail.BookInfo.BookName,
		Token:      &golabl.Task.Header.ShopMsg.Token,
		Isbn:       &detail.BookInfo.Isbn,
		GoodsID:    &detail.Detail.GoodsId,
		JSON:       &jsonStr,
		DeleteAt:   nil,
		DeleteDate: nil,
		CreateAt:   &now,
	}

	// 使用动态表名插入
	result := golabl.MysqlDb.Table(dleTaskDetailsTable).Create(delTaskDetail)
	if result.Error != nil {
		return fmt.Errorf("插入数据失败: %v", result.Error)
	}
	return nil
}

// BatchInsertDelTaskDetails 批量插入删除任务详情数据
func BatchInsertDelTaskDetails(details []planBType.DelTaskDetail) error {

	var dleTaskDetailsTable = fmt.Sprintf("del_task_details_%v", golabl.Task.TaskId)
	// 先检查并创建表
	if err := CreateTableIfNotExists(); err != nil {
		return err
	}

	if len(details) == 0 {
		return nil
	}

	now := time.Now()
	for i := range details {
		details[i].CreateAt = &now
	}

	// 批量插入，每批1000条
	batchSize := 1000
	result := golabl.MysqlDb.Table(dleTaskDetailsTable).CreateInBatches(details, batchSize)
	if result.Error != nil {
		return fmt.Errorf("批量插入数据失败: %v", result.Error)
	}
	return nil
}
