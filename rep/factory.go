package rep

import (
	"planA/initialization/golabl"
	"planA/rep/i"
	"planA/rep/impl/mysql"
	"planA/rep/impl/sqlite"
)

type Db interface {
	i.TaskExport
	i.TaskRecords
}

// CreateDbFactoryWrite 创建写数据库工厂
// @return Db mysql写数据库
// @return Db sqlite写数据库
func CreateDbFactoryWrite() (Db, Db) {
	return &mysql.GormAdapter{
			DB: golabl.MysqlDb,
		}, &sqLite.SqlAdapter{
			DB: golabl.SqliteDb,
		}
}

// CreateDbFactoryRead 创建读数据库工厂
// @return Db 读数据库
func CreateDbFactoryRead() Db {
	var read Db
	read = &mysql.GormAdapter{
		DB: golabl.MysqlDb,
	}
	if golabl.Config.Server.ReadDb == "sqlite" {
		read = &sqLite.SqlAdapter{
			DB: golabl.SqliteDb,
		}
	}
	return read
}

// CreateDbFactorySqliteRead 创建sqlite读数据库工厂
// @return Db 读数据库
func CreateDbFactorySqliteRead() Db {
	var read Db
	read = &sqLite.SqlAdapter{
		DB: golabl.SqliteDb,
	}
	return read
}
