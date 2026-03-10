package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"planA/planB/tool"
	_type "planA/planB/type"
)

// Init 初始化
func Init(ctx context.Context, config _type.MysqlConfig) (*sql.DB, error) {
	// 判断 ctx 是否取消
	checkContextErr := tool.CheckContext(ctx)
	// 判断 结果
	if checkContextErr != nil {
		// 返回 且 返回错误
		return nil, checkContextErr
	}
	// 创建数据库连接
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName))
	// 判断 错误
	if err != nil {
		// 返回 且 返回错误
		return nil, err
	}
	// 测试数据库连接
	pingErr := db.Ping()
	if pingErr != nil {
		// 关闭数据库连接
		closeErr := db.Close()
		// 判断 错误
		if closeErr != nil {
			// 返回 且 返回错误
			return nil, closeErr
		}
		// 返回 且 返回错误
		return nil, pingErr
	}
	// 返回
	return db, nil
}

// GetViolation 获取是否违规
func GetViolation(db *sql.DB, taskMsg any) (any, error) {
	//TODO

	return nil, nil
}

// IsViolation 验证是否违规
func IsViolation(violation any) bool {
	//TODO

	return false
}

// GetIsRepeat 获取是否重复
func GetIsRepeat(db *sql.DB, taskMsg any) (bool, error) {
	//TODO

	return false, nil
}
