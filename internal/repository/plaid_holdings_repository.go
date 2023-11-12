package repository

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
)

type PlaidInvestmentsHoldingsRepository interface {
	Add(tx *sql.Tx, models []model.PlaidInvestmentHoldings) error
}

type plaidInvestmentsHoldingsRepositoryHandler struct {
}

func NewPlaidInvestmentsHoldingsRepository(db *sql.DB) PlaidInvestmentsHoldingsRepository {
	return plaidInvestmentsHoldingsRepositoryHandler{}
}

func (h plaidInvestmentsHoldingsRepositoryHandler) Add(tx *sql.Tx, models []model.PlaidInvestmentHoldings) error {
	query := PlaidInvestmentHoldings.INSERT(
		PlaidInvestmentHoldings.MutableColumns,
	).MODELS(
		models,
	)

	_, err := query.Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to insert plaid investment holdings: %w", err)
	}

	return nil
}
