package db

import (
	"context"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	db_utils "hood/internal/db/utils"
)

func AddAssetsSplits(ctx context.Context, splits []*model.AssetSplit) ([]model.AssetSplit, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	t := table.AssetSplit
	stmt := t.INSERT(t.MutableColumns).
		MODELS(splits).
		ON_CONFLICT(t.Symbol, t.Ratio, t.Date).DO_NOTHING().
		RETURNING(t.AllColumns)

	result := []model.AssetSplit{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to asset asset splits: %w", err)
	}

	return result, nil
}

func AddAppliedAssetSplits(ctx context.Context, appliedAssetSplits []model.AppliedAssetSplit) ([]model.AppliedAssetSplit, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	t := table.AppliedAssetSplit
	stmt := t.INSERT(t.MutableColumns).
		MODELS(appliedAssetSplits).
		RETURNING(t.AllColumns)

	result := []model.AppliedAssetSplit{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert applied asset splits: %w", err)
	}

	return result, nil
}
