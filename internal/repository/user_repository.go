package repository

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type UserRepository interface {
	Get(userID uuid.UUID) (*model.User, error)
}

type userRepositoryHandler struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return userRepositoryHandler{
		DB: db,
	}
}

func (h userRepositoryHandler) Get(userID uuid.UUID) (*model.User, error) {
	query := User.SELECT(User.AllColumns).
		WHERE(
			User.UserID.EQ(postgres.UUID(userID)),
		)

	user := &model.User{}
	err := query.Query(h.DB, user)
	if err != nil {
		fmt.Println(query.DebugSql())
		return nil, fmt.Errorf("failed to query user id %s: %w", userID.String(), err)
	}

	return user, nil
}
