package repository

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/domain"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type PositionsRepository interface {
	Add(tx *sql.Tx, holdingsByAccountID map[uuid.UUID]domain.Holdings, source model.PositionSourceType) error
	List(tx *sql.Tx, tradingAccountID uuid.UUID) ([]domain.Position, error)
	Delete(tx *sql.Tx, tradingAccountID uuid.UUID, symbol string) error
}

type positionsRepositoryHandler struct {
}

func NewPositionsRepository(db *sql.DB) PositionsRepository {
	return positionsRepositoryHandler{}
}

func dbPositionToDomain(p model.Position) domain.Position {
	return domain.Position{
		Symbol:         p.Ticker,
		Quantity:       p.Quantity,
		TotalCostBasis: p.TotalCostBasis,
	}
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

func (h positionsRepositoryHandler) List(tx *sql.Tx, tradingAccountID uuid.UUID) ([]domain.Position, error) {
	query := Position.SELECT(Position.AllColumns).WHERE(
		postgres.AND(
			Position.TradingAccountID.EQ(postgres.UUID(tradingAccountID)),
			Position.DeletedAt.IS_NOT_NULL(),
		),
	)
	var results []model.Position
	err := query.Query(tx, &results)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Position, len(results))
	for i, p := range results {
		out[i] = dbPositionToDomain(p)
	}
	return out, nil
}

func (h positionsRepositoryHandler) Delete(tx *sql.Tx, tradingAccountID uuid.UUID, symbol string) error {
	query := Position.UPDATE(
		Position.DeletedAt,
	).SET(
		time.Now().UTC(),
	).WHERE(
		postgres.AND(
			Position.TradingAccountID.EQ(postgres.UUID(tradingAccountID)),
			Position.Ticker.EQ(postgres.String(symbol)),
			Position.DeletedAt.IS_NOT_NULL(),
		),
	)
	_, err := query.Exec(tx)
	if err != nil {
		return err
	}
	return nil
}
