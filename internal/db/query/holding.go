package db

import (
	"context"
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/view"
)

func GetVwHolding(ctx context.Context, tx *sql.Tx) ([]model.VwHolding, error) {
	v := view.VwHolding
	query := v.SELECT(v.AllColumns)

	var results []model.VwHolding
	err := query.Query(tx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
