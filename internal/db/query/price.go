package db

import (
	"context"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	db_utils "hood/internal/db/utils"
)

func AddPrices(ctx context.Context, prices []model.Price) ([]model.Price, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	t := table.Price
	stmt := t.INSERT(t.MutableColumns).
		MODELS(prices).
		RETURNING(t.AllColumns)

	result := []model.Price{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert prices: %w", err)
	}

	return result, nil
}
