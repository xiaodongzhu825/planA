package sqLite

import (
	"database/sql"
	"errors"
	"fmt"
	"planA/initialization/golabl"
	"planA/service"

	_ "modernc.org/sqlite"
)

// Init 初始化sqlIte连接
// @return error 错误信息
func Init() error {
	// 1. 打开数据库
	db, err := sql.Open("sqlite", "./taskDb.db")
	if err != nil {
		return errors.New("打开sqLite数据库失败：" + err.Error())
	}

	// 测试连接
	err = db.Ping()
	if err != nil {
		return errors.New("无法连接到sqLite数据库：" + err.Error())
	}
	golabl.SqliteDb = db
	// 自动创建表
	if err := CreateTable(); err != nil {
		return err
	}
	return nil
}

// CreateTable 自动建表
func CreateTable() error {
	createTaskIdTabErr := service.CreateTaskIdTab()
	if createTaskIdTabErr != nil {
		return fmt.Errorf("自动创建表失败: %v", createTaskIdTabErr)
	}
	createTaskExportTabErr := service.CreateTaskExportTab()
	if createTaskExportTabErr != nil {
		return fmt.Errorf("自动创建表失败: %v", createTaskExportTabErr)
	}
	return nil
}
