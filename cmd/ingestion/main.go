package main

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/service"

	"log"
	"time"

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
	defer tx.Rollback()

	tiService := service.NewTradeIngestionService()

	quantity, err := decimal.NewFromString("10.299953")
	if err != nil {
		log.Fatal(err)
	}
	price, err := decimal.NewFromString("424.77")
	if err != nil {
		log.Fatal(err)
	}
	date, err := time.Parse(time.DateOnly, "2023-10-06")
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = tiService.ProcessSellOrder(context.Background(), tx, domain.Trade{
		TradeID:     nil,
		Symbol:      "SPY",
		Quantity:    quantity,
		Price:       price,
		Date:        date,
		Description: nil,
		Custodian:   model.CustodianType_Robinhood,
		Action:      model.TradeActionType_Sell,
	})
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}
