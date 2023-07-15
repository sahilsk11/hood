package main

import (
	"context"
	"database/sql"
	price_ingestion "hood/internal/price-ingestion"
	"hood/internal/util"
	"log"
	"net/http"

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
	secrets, err := util.LoadSecrets()
	if err != nil {
		log.Fatal(err)
	}

	priceClient := price_ingestion.AlphaVantageClient{
		HttpClient: http.DefaultClient,
		ApiKey:     secrets.AlphaVantageKey,
	}

	err = price_ingestion.UpdateCurrentHoldingsPrices(ctx, priceClient)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

}
