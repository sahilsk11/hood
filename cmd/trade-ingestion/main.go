package main

import (
	"context"
	"database/sql"
	trade_ingestion "hood/internal/trade-ingestion"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgresql://postgres:postgres@localhost:5438/postgres_test?sslmode=disable"
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

	_, err = trade_ingestion.ParseTdaTransactionFile(ctx, tx, "transactions (1).csv", tiService)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

}
