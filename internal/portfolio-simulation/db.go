package portfolio_simulation

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	db_utils "hood/internal/db/utils"

	"github.com/go-jet/jet/v2/postgres"
)

func GetOpenLotsFromDb(ctx context.Context, symbol string) ([]*model.OpenLot, error) {
	tx, err := db_utils.GetTx(ctx)
	if err != nil {
		return nil, err
	}
	t := table.OpenLot
	query := t.SELECT(t.AllColumns).
		FROM(t.INNER_JOIN(
			table.Trade, table.Trade.TradeID.EQ(t.TradeID),
		)).
		WHERE(postgres.AND(
			table.Trade.Symbol.EQ(postgres.String(symbol)),
			t.DeletedAt.IS_NULL(),
		))

	var result []*model.OpenLot
	err = query.Query(tx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
