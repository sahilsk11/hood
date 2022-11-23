package db_utils

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/view"
)

func GetVwHoldings(ctx context.Context) ([]model.VwHolding, error) {
	tx, err := GetTx(ctx)
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

func GetVwOpenLotPosition(ctx context.Context) ([]model.VwOpenLotPosition, error) {
	tx, err := GetTx(ctx)
	if err != nil {
		return nil, err
	}

	v := view.VwOpenLotPosition
	query := v.SELECT(v.AllColumns)

	var results []model.VwOpenLotPosition
	err = query.Query(tx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
