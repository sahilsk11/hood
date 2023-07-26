package main

import (
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/portfolio"
	"log"

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

	custodian := model.CustodianType_Robinhood
	trades, err := db.GetHistoricTrades(tx, custodian)
	if err != nil {
		log.Fatal(err)
	}
	assetSplits, err := db.GetHistoricAssetSplits(tx)
	if err != nil {
		log.Fatal(err)
	}
	transfers, err := db.GetHistoricTransfers(tx, custodian)
	if err != nil {
		log.Fatal(err)
	}

	portfolio, err := portfolio.Playback(trades, assetSplits, transfers)
	if err != nil {
		log.Fatal(err)
	}

	symbolValue := map[string]decimal.Decimal{}
	for symbol, lots := range portfolio.OpenLots {
		symbolValue[symbol] = decimal.Zero
		for _, lot := range lots {
			symbolValue[symbol] = symbolValue[symbol].Add(lot.Quantity)
		}
	}
	fmt.Println(symbolValue)

	// tx.Commit()

}
