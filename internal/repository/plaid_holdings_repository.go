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

type PlaidInvestmentsHoldingsRepository interface {
	Add(tx *sql.Tx, holdingsByAccountID map[uuid.UUID]domain.Holdings) error
}

type plaidInvestmentsHoldingsRepositoryHandler struct {
}

func NewPlaidInvestmentsHoldingsRepository(db *sql.DB) PlaidInvestmentsHoldingsRepository {
	return plaidInvestmentsHoldingsRepositoryHandler{}
}

func holdingsByAccountIDToModels(in map[uuid.UUID]domain.Holdings) []model.PlaidInvestmentHoldings {
	out := []model.PlaidInvestmentHoldings{}
	for accountID, holdings := range in {
		for _, position := range holdings.Positions {
			out = append(out, model.PlaidInvestmentHoldings{
				Ticker:           position.Symbol,
				TradingAccountID: accountID,
				TotalCostBasis:   position.TotalCostBasis,
				Quantity:         position.Quantity,
				CreatedAt:        time.Now().UTC(),
			})
		}
	}
	return out
}

func (h plaidInvestmentsHoldingsRepositoryHandler) Add(tx *sql.Tx, holdingsByAccountID map[uuid.UUID]domain.Holdings) error {
	query := PlaidInvestmentHoldings.INSERT(
		PlaidInvestmentHoldings.MutableColumns,
	).MODELS(
		holdingsByAccountIDToModels(holdingsByAccountID),
	)

	_, err := query.Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to insert plaid investment holdings: %w", err)
	}

	return nil
}
