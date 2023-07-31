package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	hood_errors "hood/internal"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/domain"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	_ "github.com/lib/pq"
)

func AddTrades(ctx context.Context, tx *sql.Tx, dTrades []domain.Trade) ([]domain.Trade, error) {
	// will search trades for RH trades and query
	// for existing ones. This is how we loosely
	// enforce a unique constraint for only RH trades
	trades := tradesToDb(dTrades)
	err := findDuplicateRhTrades(tx, trades)
	if err != nil {
		return nil, err
	}

	stmt := Trade.INSERT(Trade.MutableColumns).
		MODELS(trades).
		RETURNING(Trade.AllColumns)

	result := []model.Trade{}
	err = stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert trades: %w", err)
	}

	return tradesFromDb(result), nil
}

func tradesFromDb(trades []model.Trade) []domain.Trade {
	out := make([]domain.Trade, len(trades))
	for i, t := range trades {
		out[i] = tradeFromDb(t)
	}
	return out
}

func GetHistoricTrades(tx *sql.Tx, custodian model.CustodianType) ([]domain.Trade, error) {
	query := Trade.SELECT(Trade.AllColumns).
		WHERE(Trade.Custodian.EQ(postgres.NewEnumValue(custodian.String()))).
		ORDER_BY(Trade.Date.ASC())
	out := []model.Trade{}
	err := query.Query(tx, &out)
	if err != nil {
		return nil, err
	}

	return tradesFromDb(out), nil
}

func findDuplicateRhTrades(tx *sql.Tx, trades []model.Trade) error {
	rhTrades := []model.Trade{}
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
				Trade.Symbol.EQ(postgres.String(t.Symbol)),
				Trade.Action.EQ(postgres.NewEnumValue(t.Action.String())),
				Trade.Quantity.EQ(postgres.Float(t.Quantity.InexactFloat64())),
				Trade.CostBasis.EQ(postgres.Float(t.CostBasis.InexactFloat64())),
				Trade.Date.EQ(postgres.TimestampzT(t.Date)),
			),
		)
	}

	query := Trade.SELECT(Trade.AllColumns).
		WHERE(postgres.AND(
			Trade.Custodian.EQ(postgres.NewEnumValue(model.CustodianType_Robinhood.String())),
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
	_, err := TdaTrade.INSERT(TdaTrade.MutableColumns).MODEL(tdaTrade).Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to add TDA trade: %w", err)
	}
	return nil
}

// adapters

func tradeToDb(t domain.Trade) model.Trade {
	return model.Trade{
		Symbol:      t.Symbol,
		Quantity:    t.Quantity,
		CostBasis:   t.Price,
		Date:        t.Date,
		Description: t.Description,
		Custodian:   t.Custodian,
		CreatedAt:   time.Now().UTC(),
		ModifiedAt:  time.Now().UTC(),
		Action:      t.Action,
	}
}

func tradeFromDb(t model.Trade) domain.Trade {
	return domain.Trade{
		TradeID:     &t.TradeID,
		Symbol:      t.Symbol,
		Quantity:    t.Quantity,
		Price:       t.CostBasis,
		Date:        t.Date,
		Description: t.Description,
		Custodian:   t.Custodian,
		Action:      t.Action,
	}
}

func tradesToDb(t []domain.Trade) []model.Trade {
	out := make([]model.Trade, len(t))
	for i, d := range t {
		out[i] = tradeToDb(d)
	}
	return out
}
