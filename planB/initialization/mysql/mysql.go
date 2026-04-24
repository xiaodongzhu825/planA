package mysql

import (
	"fmt"
	"planA/planB/initialization/golabl"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// LikeMysqlSetToG 链接mysql并保留到全局变量中
func LikeMysqlSetToG() error {

	// 1. 获取mysql配置
	mysqlConfig := golabl.Config.MysqlConfig

	// 2. 配置 DSN
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlConfig.User,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.DBName,
	)

	// 3. 配置 GORM 连接选项

	logLevel := logger.Silent
	switch mysqlConfig.Loglevel {
	case "info":
		logLevel = logger.Info
	case "warn":
		logLevel = logger.Warn
	case "error":
		logLevel = logger.Error
	case "silent":
		logLevel = logger.Silent
	}

	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logLevel), //日志级别
		DisableForeignKeyConstraintWhenMigrating: true,                             //不创建外键约束
	}

	// 4. 连接数据库
	db, openErr := gorm.Open(mysql.Open(dsn), gormConfig)
	if openErr != nil {
		return openErr
	}

	// 5. 获取底层 sql.DB，配置连接池
	sqlDB, dbErr := db.DB()
	if dbErr != nil {
		return dbErr
	}
	// 连接池优化 + 保活配置
	sqlDB.SetMaxOpenConns(mysqlConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(mysqlConfig.MaxIdleConns)
	sqlDB.SetConnMaxIdleTime(mysqlConfig.ConnMaxIdleTime * time.Minute)
	sqlDB.SetConnMaxLifetime(mysqlConfig.ConnMaxLifetime * time.Hour)

	// 5. 验证连接
	if dbPingErr := sqlDB.Ping(); dbPingErr != nil {
		return dbPingErr
	}

	// 7. 保存db实例
	golabl.MysqlDb = db
	return nil
}
