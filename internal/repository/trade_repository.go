package repository

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
)

type TradeRepository interface {
	AddPlaidMetadata(tx *sql.Tx, models []model.PlaidTradeMetadata) error
}

type tradeRepositoryHandler struct {
}

func NewTradeRepository() TradeRepository {
	return tradeRepositoryHandler{}
}

func (h tradeRepositoryHandler) AddPlaidMetadata(tx *sql.Tx, models []model.PlaidTradeMetadata) error {
	query := PlaidTradeMetadata.INSERT(
		PlaidTradeMetadata.MutableColumns,
	).MODELS(models)

	_, err := query.Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to add plaid trade metadata: %w", err)
	}

	return nil
}
