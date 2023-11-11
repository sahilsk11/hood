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
	Add(tx *sql.Tx, userID uuid.UUID, plaidItemID, accessToken string) (*model.PlaidItem, error)
}

type plaidItemRepositoryHandler struct {
	DB *sql.DB
}

func NewPlaidItemRepository(db *sql.DB) PlaidItemRepository {
	return plaidItemRepositoryHandler{
		DB: db,
	}
}

func (h plaidItemRepositoryHandler) Add(tx *sql.Tx, userID uuid.UUID, plaidItemID, accessToken string) (*model.PlaidItem, error) {
	// TODO - update migration so we use generated uuid
	query := PlaidItem.INSERT(
		PlaidItem.AllColumns,
	).MODEL(
		model.PlaidItem{
			ItemID:      uuid.New(),
			PlaidItemID: plaidItemID,
			AccessToken: accessToken,
			CreatedAt:   time.Now().UTC(),
			UserID:      userID,
		},
	).RETURNING(
		PlaidItem.AllColumns,
	)

	plaidItem := &model.PlaidItem{}
	err := query.Query(tx, plaidItem)
	if err != nil {
		return nil, fmt.Errorf("failed to insert item for %s: %w", userID.String(), err)
	}

	return plaidItem, nil
}
