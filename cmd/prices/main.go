package main

import (
	db "hood/internal/db/query"
	"hood/internal/prices"
	"hood/internal/util"
	"log"
	"net/http"
	"time"
)

func main() {
	tx, err := db.NewTx()
	if err != nil {
		log.Fatal(err)
	}

	secrets, err := util.LoadSecrets()
	if err != nil {
		log.Fatal(err)
	}

	priceClient := prices.AlphaVantageClient{
		HttpClient: http.DefaultClient,
		ApiKey:     secrets.AlphaVantageKey,
	}

	err = prices.UpdateHistoric(tx, priceClient, []string{
		"AAPL", "COIN", "AMZN", "NVDA",
	}, time.Now().AddDate(0, 0, -14))
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}
