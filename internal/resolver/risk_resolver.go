package resolver

import (
	"database/sql"
)

type Resolver struct {
	Db *sql.DB
}
