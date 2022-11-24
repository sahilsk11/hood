package db

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/view"
	db_utils "hood/internal/db/utils"
	"hood/internal/domain"
)

func GetOpenLots(ctx context.Context, symbol string) ([]*domain.OpenLot, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT open_lot.open_lot_id, trade.trade_id, trade.symbol, open_lot.quantity, open_lot.cost_basis, trade.date AS "purchase_date"
	FROM open_lot
	INNER JOIN trade on trade.trade_id = open_lot.trade_id
	WHERE trade.symbol = $1 AND deleted_at is null
	ORDER BY "purchase_date";
	`
	rows, err := tx.QueryContext(ctx, query, symbol)
	if err != nil {
		return nil, err
	}

	var result []*domain.OpenLot
	for rows.Next() {
		openLot := domain.OpenLot{}
		err = rows.Scan(
			&openLot.OpenLotID,
			&openLot.TradeID,
			&openLot.Symbol,
			&openLot.Quantity,
			&openLot.CostBasis,
			&openLot.PurchaseDate,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, &openLot)
	}

	return result, nil
}

func GetVwOpenLotPosition(ctx context.Context) ([]model.VwOpenLotPosition, error) {
	tx, err := db_utils.GetTx(ctx)
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
