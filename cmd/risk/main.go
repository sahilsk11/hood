package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/metrics"
	"hood/internal/portfolio"
	"io/ioutil"
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

	fmt.Println(metrics.StdevOfAsset(tx, "NVDA"))

	// Read values from JSON file
	values := make(map[string]decimal.Decimal)
	valuesFile, err := ioutil.ReadFile("values.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(valuesFile, &values)
	if err != nil {
		log.Fatal(err)
	}

	// Read transferMap from JSON file
	transfersMap := make(map[string]decimal.Decimal)
	transferMapFile, err := ioutil.ReadFile("transfersMap.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(transferMapFile, &transfersMap)
	if err != nil {
		log.Fatal(err)
	}

	// Write values to JSON file
	valuesJson, err := json.Marshal(values)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("values.json", valuesJson, 0644)
	if err != nil {
		log.Fatal(err)
	}

	out, err := metrics.TimeWeightedReturns(values, transfersMap)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(metrics.StandardDeviation(tx, out))

}

func getData(tx *sql.Tx, custodian model.CustodianType) portfolio.Events {
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
	dividends, err := db.GetHistoricDividends(tx, custodian)
	if err != nil {
		log.Fatal(err)
	}

	return portfolio.Events{
		Trades:      trades,
		AssetSplits: assetSplits,
		Transfers:   transfers,
		Dividends:   dividends,
	}
}
