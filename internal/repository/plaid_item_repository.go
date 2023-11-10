package repository

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"time"

	"github.com/google/uuid"
)

type PlaidItemRepository interface {
	Add(userID uuid.UUID, plaidItemID, accessToken string) (*model.PlaidItem, error)
}

type plaidItemRepositoryHandler struct {
	DB *sql.DB
}

func NewPlaidItemRepository(db *sql.DB) PlaidItemRepository {
	return plaidItemRepositoryHandler{
		DB: db,
	}
}

func (h plaidItemRepositoryHandler) Add(userID uuid.UUID, plaidItemID, accessToken string) (*model.PlaidItem, error) {
	query := PlaidItem.INSERT(
		PlaidItem.MutableColumns,
	).MODEL(
		model.PlaidItem{
			PlaidItemID: plaidItemID,
			AccessToken: accessToken,
			CreatedAt:   time.Now().UTC(),
			UserID:      userID,
		},
	)

	plaidItem := &model.PlaidItem{}
	err := query.Query(h.DB, plaidItem)
	if err != nil {
		return nil, fmt.Errorf("failed to insert item for %s: %w", userID.String(), err)
	}

	return plaidItem, nil
}
