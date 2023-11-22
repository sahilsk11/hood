package repository

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/domain"
	"time"

	"github.com/google/uuid"
)

type PositionsRepository interface {
	Add(tx *sql.Tx, holdingsByAccountID map[uuid.UUID]domain.Holdings, source model.PositionSourceType) error
}

type positionsRepositoryHandler struct {
}

func NewPositionsRepository(db *sql.DB) PositionsRepository {
	return positionsRepositoryHandler{}
}

func holdingsByAccountIDToModels(in map[uuid.UUID]domain.Holdings, source model.PositionSourceType) []model.Position {
	out := []model.Position{}
	for accountID, holdings := range in {
		for _, position := range holdings.Positions {
			out = append(out, model.Position{
				Ticker:           position.Symbol,
				TradingAccountID: accountID,
				TotalCostBasis:   position.TotalCostBasis,
				Quantity:         position.Quantity,
				CreatedAt:        time.Now().UTC(),
				Source:           source,
			})
		}
	}
	return out
}

func (h positionsRepositoryHandler) Add(tx *sql.Tx, holdingsByAccountID map[uuid.UUID]domain.Holdings, source model.PositionSourceType) error {
	query := Position.INSERT(
		Position.MutableColumns,
	).MODELS(
		holdingsByAccountIDToModels(holdingsByAccountID, source),
	)

	_, err := query.Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to insert plaid investment holdings: %w", err)
	}

	return nil
}
