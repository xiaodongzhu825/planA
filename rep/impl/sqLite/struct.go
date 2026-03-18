package sqLite

import "database/sql"

type SqlAdapter struct {
	DB *sql.DB
}
