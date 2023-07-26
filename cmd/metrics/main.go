package main

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/metrics"
	"hood/internal/portfolio"
	"log"
	"sort"
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

	e1 := getData(tx, model.CustodianType_Robinhood)
	// e2 := getData(tx, model.CustodianType_Robinhood)
	// events := portfolio.Events{
	// 	Trades:      append(e1.Trades, e2.Trades...),
	// 	AssetSplits: append(e1.AssetSplits, e2.AssetSplits...),
	// 	Transfers:   append(e1.Transfers, e2.Transfers...),
	// 	Dividends:   append(e1.Dividends, e2.Dividends...),
	// }
	events := e1
	dailyPortfolio, err := portfolio.PlaybackDaily(events)
	if err != nil {
		log.Fatal(err)
	}
	transfers := e1.Transfers //append(e1.Transfers, e2.Transfers...)
	tranfersMap := map[string]decimal.Decimal{}
	for _, t := range transfers {
		dateStr := t.Date.Format("2006-01-02")
		if _, ok := tranfersMap[dateStr]; !ok {
			tranfersMap[dateStr] = decimal.Zero
		}
		tranfersMap[dateStr] = tranfersMap[dateStr].Add(t.Amount)
	}
	start, err := time.Parse("2006-01-02", "2020-06-19")
	if err != nil {
		log.Fatal(err)
	}
	end, err := time.Parse("2006-01-02", "2023-07-15")
	if err != nil {
		log.Fatal(err)
	}
	values, err := metrics.DailyPortfolioValues(
		tx,
		dailyPortfolio,
		&start,
		&end,
	)
	// fmt.Println(values)

	if err != nil {
		log.Fatal(err)
	}

	out, err := metrics.TimeWeightedReturns(values, tranfersMap)
	if err != nil {
		log.Fatal(err)
	}

	dateKeys := []string{}
	for d := range out {
		dateKeys = append(dateKeys, d)
	}
	sort.Strings(dateKeys)
	for _, d := range dateKeys {
		fmt.Printf("%s,%f\n", d, out[d].InexactFloat64())
	}
	// fmt.Println(out[dateKeys[len(dateKeys)-1]].String())

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
