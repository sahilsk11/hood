package db

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/db/models/postgres/public/view"
	"hood/internal/domain"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
)

func GetOpenLots(ctx context.Context, tx *sql.Tx, symbol string, custodian model.CustodianType) ([]domain.OpenLot, error) {
	result := []struct {
		model.OpenLot
		model.Trade
	}{}
	query := OpenLot.SELECT(OpenLot.AllColumns, Trade.AllColumns).FROM(
		OpenLot.INNER_JOIN(Trade, Trade.TradeID.EQ(OpenLot.TradeID)),
	).WHERE(
		AND(
			OpenLot.Quantity.GT(Float(0)),
			Trade.Symbol.EQ(String(symbol)),
			Trade.Custodian.EQ(NewEnumValue(custodian.String())),
		),
	).ORDER_BY(OpenLot.Date.ASC())
	err := query.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get open lots from db: %w", err)
	}

	out := []domain.OpenLot{}
	for _, r := range result {
		lot := openLotFromDb(r.OpenLot, r.Trade)
		out = append(out, lot)
	}
	return out, nil
}

func openLotFromDb(o model.OpenLot, t model.Trade) domain.OpenLot {
	return domain.OpenLot{
		Trade:     tradeFromDb(t).Ptr(),
		OpenLotID: &o.OpenLotID,
		LotID:     o.LotID,
		Quantity:  o.Quantity,
		CostBasis: o.CostBasis,
		Date:      o.Date,
	}
}

func GetVwOpenLotPosition(ctx context.Context, tx *sql.Tx) ([]model.VwOpenLotPosition, error) {

	v := view.VwOpenLotPosition
	query := v.SELECT(v.AllColumns)

	var results []model.VwOpenLotPosition
	err := query.Query(tx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func AddOpenLots(ctx context.Context, tx *sql.Tx, openLots []domain.OpenLot) ([]domain.OpenLot, error) {
	result := []struct {
		model.OpenLot
		model.Trade
	}{}
	t := OpenLot
	add := CTE("add_open_lots")
	tradeID := OpenLot.TradeID.From(add)

	stmt := WITH(
		add.AS(
			t.INSERT(t.MutableColumns).
				MODELS(openLotsToDb(openLots)).
				RETURNING(t.AllColumns),
		),
	)(
		SELECT(add.AllColumns(), Trade.AllColumns).
			FROM(
				add.INNER_JOIN(Trade, Trade.TradeID.EQ(tradeID)),
			),
	)

	err := stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to add open lots to db: %w", err)
	}
	out := []domain.OpenLot{}
	for _, r := range result {
		lot := openLotFromDb(r.OpenLot, r.Trade)
		out = append(out, lot)
	}

	return out, nil
}

func openLotsToDb(lots []domain.OpenLot) []model.OpenLot {
	out := make([]model.OpenLot, len(lots))
	for i, l := range lots {
		out[i] = model.OpenLot{
			CostBasis:  l.CostBasis,
			Quantity:   l.Quantity,
			TradeID:    *l.TradeID(),
			CreatedAt:  time.Now().UTC(),
			ModifiedAt: time.Now().UTC(),
			LotID:      l.LotID,
			Date:       l.Date,
		}
	}
	return out
}

func AddClosedLots(ctx context.Context, tx *sql.Tx, lots []domain.ClosedLot) ([]*model.ClosedLot, error) {
	// TODO: validate input
	t := ClosedLot
	stmt := t.INSERT(t.MutableColumns).
		MODELS(closedLotsToDb(lots)).
		RETURNING(t.AllColumns)

	result := []*model.ClosedLot{}
	err := stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert closed lots: %w", err)
	}

	return result, nil
}

func closedLotsToDb(lots []domain.ClosedLot) []model.ClosedLot {
	out := make([]model.ClosedLot, len(lots))
	for i, lot := range lots {
		out[i] = model.ClosedLot{
			BuyTradeID:    *lot.OpenLot.TradeID(),
			SellTradeID:   *lot.SellTrade.TradeID,
			Quantity:      lot.Quantity,
			RealizedGains: lot.RealizedGains,
			GainsType:     lot.GainsType,
			CreatedAt:     time.Now().UTC(),
			ModifiedAt:    time.Now().UTC(),
		}
	}
	return out
}

func UpdateOpenLotInDb(ctx context.Context, tx *sql.Tx, updatedLot model.OpenLot, columns ColumnList) (*model.OpenLot, error) {
	t := OpenLot
	updatedLot.ModifiedAt = time.Now().UTC()
	columns = append(columns, t.ModifiedAt)

	stmt := t.UPDATE(columns).
		MODEL(updatedLot).
		WHERE(t.OpenLotID.EQ(Int32(updatedLot.OpenLotID))).
		RETURNING(t.AllColumns)

	result := model.OpenLot{}
	err := stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert closed lots: %w", err)
	}

	return &result, nil
}
