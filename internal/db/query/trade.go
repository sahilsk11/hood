package db

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/domain"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func tradesFromDb(trades []model.Trade) []domain.Trade {
	out := make([]domain.Trade, len(trades))
	for i, t := range trades {
		out[i] = tradeFromDb(t)
	}
	return out
}

func GetHistoricTrades(tx *sql.Tx, tradingAccountID uuid.UUID) ([]domain.Trade, error) {
	query := Trade.SELECT(Trade.AllColumns).
		WHERE(Trade.TradingAccountID.EQ(postgres.String(tradingAccountID.String()))).
		ORDER_BY(Trade.Date.ASC())
	out := []model.Trade{}
	err := query.Query(tx, &out)
	if err != nil {
		return nil, err
	}

	return tradesFromDb(out), nil
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
		Symbol:           t.Symbol,
		Quantity:         t.Quantity,
		CostBasis:        t.Price,
		Date:             t.Date,
		Description:      t.Description,
		TradingAccountID: t.TradingAccountID,
		CreatedAt:        time.Now().UTC(),
		ModifiedAt:       time.Now().UTC(),
		Action:           t.Action,
	}
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
	}
}

func tradesToDb(t []domain.Trade) []model.Trade {
	out := make([]model.Trade, len(t))
	for i, d := range t {
		out[i] = tradeToDb(d)
	}
	return out
}
