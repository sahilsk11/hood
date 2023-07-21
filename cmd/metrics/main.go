package main

import (
	"fmt"
	db "hood/internal/db/query"
	metrics "hood/internal/metrics"
	"hood/internal/util"
	"log"
	"sort"
	"time"

	_ "github.com/lib/pq"
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

	trades, err := db.GetHistoricTrades(tx)
	if err != nil {
		log.Fatal(err)
	}
	assetSplits, err := db.GetHistoricAssetSplits(tx)
	if err != nil {
		log.Fatal(err)
	}
	startTime := time.Date(2020, 06, 19, 0, 0, 0, 0, time.UTC)
	endTime := time.Now().UTC() //time.Date(2020, 06, 24, 0, 0, 0, 0, time.UTC)

	transfers, err := db.GetHistoricTransfers(tx)
	if err != nil {
		log.Fatal(err)
	}
	transfersMap := map[string]decimal.Decimal{}
	for _, t := range transfers {
		dateStr := t.Date.Format("2006-01-02")
		if _, ok := transfersMap[dateStr]; !ok {
			transfersMap[dateStr] = decimal.Zero
		}
		transfersMap[dateStr] = transfersMap[dateStr].Add(t.Amount)
	}
	util.Pprint(transfers[:5])
	util.EnableDebug = false
	dailyPortfolio, err := metrics.CalculateDailyPortfolios(trades, assetSplits, transfers, startTime, endTime)
	if err != nil {
		log.Fatal(err)
	}
	util.Pprint(dailyPortfolio)

	portfolioValues, err := metrics.CalculateNetPortfolioValues(tx, dailyPortfolio)
	if err != nil {
		log.Fatal(err)
	}

	util.Pprint(portfolioValues)

	// Pprint(portfolioValues)

	out, err := metrics.TimeWeightedReturns(portfolioValues, transfersMap)
	if err != nil {
		log.Fatal(err)
	}

	// bytes, err := json.MarshalIndent(out, "", "    ")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(bytes))

	dates := []string{}
	for k := range out {
		dates = append(dates, k)
	}

	sort.Strings(dates)
	for _, v := range dates {
		fmt.Printf("%s,%f\n", v, out[v].InexactFloat64())
	}

	// bytes, err = json.MarshalIndent(dailyPortfolio, "", "    ")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(bytes))

	// fmt.Println(metrics.CalculateNetReturns(tx))
}
