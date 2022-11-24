package db

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	db_utils "hood/internal/db/utils"

	"github.com/go-jet/jet/v2/postgres"
)

func AddTradesToDb(ctx context.Context, trades []*model.Trade) ([]model.Trade, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	stmt := table.Trade.INSERT(table.Trade.MutableColumns).
		MODELS(trades).
		ON_CONFLICT(
			table.Trade.Symbol,
			table.Trade.Action,
			table.Trade.Quantity,
			table.Trade.CostBasis,
			table.Trade.Date,
		).DO_UPDATE(
		postgres.SET(
			table.Trade.Symbol.SET(table.Trade.EXCLUDED.Symbol),
		),
	).
		RETURNING(table.Trade.AllColumns)

	result := []model.Trade{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
