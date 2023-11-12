package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/domain"
	"time"

	"github.com/go-jet/jet/v2/postgres"

	hood_errors "hood/internal"
)

type TradeRepository interface {
	Add(tx *sql.Tx, dTrades []domain.Trade) ([]domain.Trade, error)
	AddPlaidMetadata(tx *sql.Tx, models []model.PlaidTradeMetadata) error
}

type tradeRepositoryHandler struct {
}

func NewTradeRepository() TradeRepository {
	return tradeRepositoryHandler{}
}

func (h tradeRepositoryHandler) Add(tx *sql.Tx, dTrades []domain.Trade) ([]domain.Trade, error) {
	// will search trades for RH trades and query
	// for existing ones. This is how we loosely
	// enforce a unique constraint for only RH trades
	trades := tradesToDb(dTrades)
	err := findDuplicateTrades(tx, trades)
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

func tradeFromDb(t model.Trade) domain.Trade {
	return domain.Trade{
		TradeID:          &t.TradeID,
		Symbol:           t.Symbol,
		Quantity:         t.Quantity,
		Price:            t.CostBasis,
		Date:             t.Date,
		Description:      t.Description,
		TradingAccountID: t.TradingAccountID,
		Action:           t.Action,
		Source:           t.Source,
	}
}

func tradesToDb(t []domain.Trade) []model.Trade {
	out := make([]model.Trade, len(t))
	for i, d := range t {
		out[i] = tradeToDb(d)
	}
	return out
}

func tradeToDb(t domain.Trade) model.Trade {
	return model.Trade{
		Symbol:           t.Symbol,
		Quantity:         t.Quantity,
		CostBasis:        t.Price,
		Date:             t.Date,
		Description:      t.Description,
		TradingAccountID: t.TradingAccountID,
		CreatedAt:        time.Now().UTC(),
		ModifiedAt:       time.Now().UTC(),
		Action:           t.Action,
		Source:           t.Source,
	}
}

func tradesFromDb(trades []model.Trade) []domain.Trade {
	out := make([]domain.Trade, len(trades))
	for i, t := range trades {
		out[i] = tradeFromDb(t)
	}
	return out
}

func (h tradeRepositoryHandler) AddPlaidMetadata(tx *sql.Tx, models []model.PlaidTradeMetadata) error {
	query := PlaidTradeMetadata.INSERT(
		PlaidTradeMetadata.MutableColumns,
	).MODELS(models)

	_, err := query.Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to add plaid trade metadata: %w", err)
	}

	return nil
}

// TODO - think about how to handle this
func findDuplicateTrades(tx *sql.Tx, trades []model.Trade) error {

	if len(trades) == 0 {
		return nil
	}

	exp := []postgres.BoolExpression{}
	for _, t := range trades {
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
			postgres.OR(exp...),
		))

	result := []model.Trade{}
	err := query.Query(tx, &result)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check for existing trades: %w", err)
	}

	if len(result) > 0 {
		return hood_errors.ErrDuplicateTrade{
			// Custodian: mod, //idk man
			Message: fmt.Sprintf("%v", result[0]),
		}
	}

	return nil
}
