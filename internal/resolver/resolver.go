package resolver

import (
	"database/sql"
	"hood/internal/repository"
)

type Resolver struct {
	Db                  *sql.DB
	PlaidRepository     repository.PlaidRepository
	UserRepository      repository.UserRepository
	PlaidItemRepository repository.PlaidItemRepository
}
