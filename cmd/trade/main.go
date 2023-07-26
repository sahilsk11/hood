package main

import (
	db "hood/internal/db/query"
	"hood/internal/portfolio"
	"log"
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
	transfers, err := db.GetHistoricTransfers(tx)
	if err != nil {
		log.Fatal(err)
	}

	err = portfolio.Playback(trades, assetSplits, transfers)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

}
