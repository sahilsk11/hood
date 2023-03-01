package db

import (
	"context"
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
)

func AddTrades(ctx context.Context, tx *sql.Tx, trades []*model.Trade) ([]model.Trade, error) {
	stmt := table.Trade.INSERT(table.Trade.MutableColumns).
		MODELS(trades).
		ON_CONFLICT(
			table.Trade.Symbol,
			table.Trade.Action,
			table.Trade.Quantity,
			table.Trade.CostBasis,
			table.Trade.Date,
		).DO_NOTHING().
		RETURNING(table.Trade.AllColumns)

	result := []model.Trade{}
	err := stmt.Query(tx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
