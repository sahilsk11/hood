package main

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	portfolio_simulation "hood/internal/portfolio-simulation"
	"log"

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

	result, err := portfolio_simulation.SimulateTrade(ctx, model.Trade{
		Symbol:    "ATLASSIAN",
		Quantity:  decimal.NewFromFloat(0.681328),
		Action:    model.TradeActionType_Sell,
		CostBasis: decimal.NewFromFloat(118.7),
	})
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Rollback()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
