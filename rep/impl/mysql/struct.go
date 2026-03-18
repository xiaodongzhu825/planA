package mysql

import (
	"gorm.io/gorm"
)

type GormAdapter struct {
	DB *gorm.DB
}
