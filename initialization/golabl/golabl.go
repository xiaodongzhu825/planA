package golabl

import (
	"context"
	"database/sql"
	_type "planA/type"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var (
	Ctx      context.Context
	Config   _type.Config
	MysqlDb  *gorm.DB
	RedisDbA *redis.Client
	RedisDbB *redis.Client
	RedisDbC *redis.Client
	SqliteDb *sql.DB
	Router   = mux.NewRouter()
)
