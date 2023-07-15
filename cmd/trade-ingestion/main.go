package main

import (
	"context"
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	trade_ingestion "hood/internal/trade-ingestion"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func main() {
	connStr := "postgresql://postgres:postgres@localhost:5438/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.WithValue(
		context.Background(),
		"tx",
		tx,
	)

	tiService := trade_ingestion.NewTradeIngestionService(ctx, tx)

	_, _, err = tiService.ProcessSellOrder(ctx, tx, model.Trade{
		Symbol:     "SPY",
		Action:     model.TradeActionType_Sell,
		Quantity:   decimal.NewFromFloat(12.279503),
		CostBasis:  decimal.NewFromFloat(407.18),
		Date:       time.Date(2023, 03, 31, 6, 12, 20, 0, time.Local),
		CreatedAt:  time.Now().UTC(),
		ModifiedAt: time.Now().UTC(),
		Custodian:  model.CustodianType_Robinhood,
	})
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

}
