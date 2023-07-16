package db

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	. "hood/internal/db/models/postgres/public/table"
)

func AddAssetsSplits(ctx context.Context, tx *sql.Tx, splits []*model.AssetSplit) ([]model.AssetSplit, error) {
	t := AssetSplit
	stmt := t.INSERT(t.MutableColumns).
		MODELS(splits).
		ON_CONFLICT(t.Symbol, t.Ratio, t.Date).DO_NOTHING().
		RETURNING(t.AllColumns)

	result := []model.AssetSplit{}
	err := stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to asset asset splits: %w", err)
	}

	return result, nil
}

func AddAppliedAssetSplits(ctx context.Context, tx *sql.Tx, appliedAssetSplits []model.AppliedAssetSplit) ([]model.AppliedAssetSplit, error) {

	t := table.AppliedAssetSplit
	stmt := t.INSERT(t.MutableColumns).
		MODELS(appliedAssetSplits).
		RETURNING(t.AllColumns)

	result := []model.AppliedAssetSplit{}
	err := stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert applied asset splits: %w", err)
	}

	return result, nil
}

func GetHistoricAssetSplits(tx *sql.Tx) ([]model.AssetSplit, error) {
	query := AssetSplit.SELECT(AssetSplit.AllColumns).
		ORDER_BY(AssetSplit.Date.ASC())
	out := []model.AssetSplit{}
	err := query.Query(tx, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
