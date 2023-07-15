package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	hood_errors "hood/internal"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"

	"github.com/go-jet/jet/v2/postgres"
	_ "github.com/lib/pq"
)

func AddTrades(ctx context.Context, tx *sql.Tx, trades []*model.Trade) ([]model.Trade, error) {
	// will search trades for RH trades and query
	// for existing ones. This is how we loosely
	// enforce a unique constraint for only RH trades
	err := findDuplicateRhTrades(tx, trades)
	if err != nil {
		return nil, err
	}

	stmt := table.Trade.INSERT(table.Trade.MutableColumns).
		MODELS(trades).
		RETURNING(table.Trade.AllColumns)

	result := []model.Trade{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func findDuplicateRhTrades(tx *sql.Tx, trades []*model.Trade) error {
	rhTrades := []*model.Trade{}
	for _, t := range trades {
		if t.Custodian == model.CustodianType_Robinhood {
			rhTrades = append(rhTrades, t)
		}
	}

	if len(rhTrades) == 0 {
		return nil
	}

	exp := []postgres.BoolExpression{}
	for _, t := range rhTrades {
		exp = append(
			exp,
			postgres.AND(
				table.Trade.Symbol.EQ(postgres.String(t.Symbol)),
				table.Trade.Action.EQ(postgres.NewEnumValue(t.Action.String())),
				table.Trade.Quantity.EQ(postgres.Float(t.Quantity.InexactFloat64())),
				table.Trade.CostBasis.EQ(postgres.Float(t.CostBasis.InexactFloat64())),
				table.Trade.Date.EQ(postgres.TimestampzT(t.Date)),
			),
		)
	}

	query := table.Trade.SELECT(table.Trade.AllColumns).
		WHERE(postgres.AND(
			table.Trade.Custodian.EQ(postgres.NewEnumValue(model.CustodianType_Robinhood.String())),
			postgres.OR(exp...),
		))

	result := []model.Trade{}
	err := query.Query(tx, &result)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check for existing RH trades: %w", err)
	}

	if len(result) > 0 {
		return hood_errors.ErrDuplicateTrade{
			Custodian: model.CustodianType_Robinhood,
			Message:   fmt.Sprintf("%v", result[0]),
		}
	}

	return nil
}

func AddTdaTrade(tx *sql.Tx, tdaTrade model.TdaTrade) error {
	_, err := table.TdaTrade.INSERT(table.TdaTrade.MutableColumns).MODEL(tdaTrade).Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to add TDA trade: %w", err)
	}
	return nil
}
