package data_ingestion

import (
	"context"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	db_utils "hood/internal/db/utils"
	"time"

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

func AddOpenLotsToDb(ctx context.Context, openLots []*model.OpenLot) ([]model.OpenLot, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	t := table.OpenLot
	stmt := t.INSERT(t.MutableColumns).
		MODELS(openLots).
		ON_CONFLICT(t.TradeID).DO_UPDATE(
		postgres.SET(t.TradeID.SET(t.EXCLUDED.TradeID)),
	).
		RETURNING(t.AllColumns)

	result := []model.OpenLot{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetOpenLotsFromDb(ctx context.Context) ([]*model.OpenLot, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}
	t := table.OpenLot
	query := t.SELECT(t.AllColumns)

	result := []*model.OpenLot{}
	err = query.Query(tx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func AddAssetsSplitsToDb(ctx context.Context, splits []*model.AssetSplit) ([]model.AssetSplit, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	t := table.AssetSplit
	stmt := t.INSERT(t.MutableColumns).
		MODELS(splits).
		ON_CONFLICT(t.Symbol, t.Ratio, t.Date).DO_UPDATE(
		postgres.SET(t.Symbol.SET(t.EXCLUDED.Symbol)),
	).
		RETURNING(t.AllColumns)

	result := []model.AssetSplit{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to asset asset splits: %w", err)
	}

	return result, nil
}

func AddAppliedAssetSplitsToDb(ctx context.Context, appliedAssetSplits []model.AppliedAssetSplit) ([]model.AppliedAssetSplit, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	t := table.AppliedAssetSplit
	stmt := t.INSERT(t.MutableColumns).
		MODELS(appliedAssetSplits).
		ON_CONFLICT(t.AssetSplitID, t.OpenLotID).DO_UPDATE(
		postgres.SET(
			t.AppliedAssetSplitID.SET(t.EXCLUDED.AppliedAssetSplitID),
		),
	).
		RETURNING(t.AllColumns)

	result := []model.AppliedAssetSplit{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert applied asset splits: %w", err)
	}

	return result, nil
}

func AddPricesToDb(ctx context.Context, prices []model.Price) ([]model.Price, error) {
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

func AddClosedLotsToDb(ctx context.Context, lots []*model.ClosedLot) ([]*model.ClosedLot, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	t := table.ClosedLot
	stmt := t.INSERT(t.MutableColumns).
		MODELS(lots).
		ON_CONFLICT(t.BuyTradeID, t.SellTradeID).
		DO_UPDATE(
			postgres.SET(
				t.BuyTradeID.SET(t.EXCLUDED.BuyTradeID),
			),
		).
		RETURNING(t.AllColumns)

	result := []*model.ClosedLot{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert closed lots: %w", err)
	}

	return result, nil
}

func UpdateOpenLotInDb(ctx context.Context, updatedLot model.OpenLot, columns postgres.ColumnList) (*model.OpenLot, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	t := table.OpenLot
	updatedLot.ModifiedAt = time.Now().UTC()
	columns = append(columns, t.ModifiedAt)

	stmt := t.UPDATE(columns).
		MODEL(updatedLot).
		RETURNING(t.AllColumns)

	result := &model.OpenLot{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert closed lots: %w", err)
	}

	return result, nil
}
