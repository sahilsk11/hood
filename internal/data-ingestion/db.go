package data_ingestion

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"

	"github.com/go-jet/jet/v2/postgres"
)

type Deps struct {
	Db *sql.DB
}

func (m Deps) AddTradesToDb(trades []*model.Trade) ([]model.Trade, error) {
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
	err := stmt.Query(m.Db, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m Deps) AddOpenLotsToDb(openLots []*model.OpenLot) ([]model.OpenLot, error) {
	t := table.OpenLot
	stmt := t.INSERT(t.MutableColumns).
		MODELS(openLots).
		ON_CONFLICT(t.TradeID).DO_UPDATE(
		postgres.SET(t.TradeID.SET(t.EXCLUDED.TradeID)),
	).
		RETURNING(t.AllColumns)

	result := []model.OpenLot{}
	err := stmt.Query(m.Db, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m Deps) AddAssetsSplitsToDb(splits []*model.AssetSplit) ([]model.AssetSplit, error) {
	t := table.AssetSplit
	stmt := t.INSERT(t.MutableColumns).
		MODELS(splits).
		ON_CONFLICT(t.Symbol, t.Ratio, t.Date).DO_UPDATE(
		postgres.SET(t.Symbol.SET(t.EXCLUDED.Symbol)),
	).
		RETURNING(t.AllColumns)

	result := []model.AssetSplit{}
	err := stmt.Query(m.Db, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to asset asset splits: %w", err)
	}

	return result, nil
}

func (m Deps) AddAppliedAssetSplitsToDb(appliedAssetSplits []model.AppliedAssetSplit) ([]model.AppliedAssetSplit, error) {
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
	err := stmt.Query(m.Db, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert applied asset splits: %w", err)
	}

	return result, nil
}

func (m Deps) AddPricesToDb(prices []model.Price) ([]model.Price, error) {
	t := table.Price
	stmt := t.INSERT(t.MutableColumns).
		MODELS(prices).
		RETURNING(t.AllColumns)

	result := []model.Price{}
	err := stmt.Query(m.Db, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert prices: %w", err)
	}

	return result, nil
}

func (m Deps) AddClosedLotsToDb(lots []*model.ClosedLot) ([]model.ClosedLot, error) {
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

	result := []model.ClosedLot{}
	err := stmt.Query(m.Db, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert closed lots: %w", err)
	}

	return result, nil
}
