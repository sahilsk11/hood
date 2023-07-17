package main

import (
	"encoding/json"
	"fmt"
	db "hood/internal/db/query"
	metrics "hood/internal/metrics"
	"log"
	"time"

	_ "github.com/lib/pq"
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

	trades, err := db.GetHistoricTrades(tx)
	if err != nil {
		log.Fatal(err)
	}
	assetSplits, err := db.GetHistoricAssetSplits(tx)
	if err != nil {
		log.Fatal(err)
	}
	startTime := time.Date(2020, 06, 18, 0, 0, 0, 0, time.UTC)
	endTime := time.Now()

	dailyPortfolio, err := metrics.DailyPortfolio(trades, assetSplits, startTime, endTime)
	if err != nil {
		log.Fatal(err)
	}

	out, err := metrics.DailyReturns(tx, dailyPortfolio)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(bytes))

	fmt.Println(metrics.CalculateNetReturns(tx))
}
