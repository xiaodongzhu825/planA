package golabl

import (
	"context"
	"database/sql"
	_type "planA/type"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var (
	Ctx       context.Context
	Config    _type.Config
	MysqlDb   *gorm.DB
	RedisDbA  *redis.Client
	RedisDbB  *redis.Client
	RedisDbC  *redis.Client
	RedisDbD  *redis.Client
	SqliteDb  *sql.DB
	Router    = mux.NewRouter()
	RedisExp  = time.Duration(Config.Server.RedisExp) * time.Hour
	Validator *validator.Validate
)
