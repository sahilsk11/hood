package main

import (
	"context"
	"database/sql"
	trade_ingestion "hood/internal/trade-ingestion"
	"log"

	_ "github.com/lib/pq"
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

	_, err = trade_ingestion.ParseTdaTransactionFile(ctx, "transactions.csv")
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

}
