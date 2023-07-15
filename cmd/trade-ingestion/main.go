package main

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	trade "hood/internal/trade"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func main() {
	dbConn, err := db.New()
	if err != nil {
		log.Fatal(err)
	}

	tx, err := dbConn.Begin()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.WithValue(
		context.Background(),
		"tx",
		tx,
	)

	tiService := trade.NewTradeIngestionService()

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
