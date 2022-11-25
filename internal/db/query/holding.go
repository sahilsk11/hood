package db

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/view"
	db_utils "hood/internal/db/utils"
)

func GetVwHolding(ctx context.Context) ([]model.VwHolding, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	v := view.VwHolding
	query := v.SELECT(v.AllColumns)

	var results []model.VwHolding
	err = query.Query(tx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
