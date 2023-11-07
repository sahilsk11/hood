package main

import (
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/metrics"
	"hood/internal/service"
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

	p, err := service.GetAggregatePortfolio(tx)
	if err != nil {
		log.Fatal(err)
	}

	// portfolioStdev, err := metrics.DailyStdevOfPortfolio(tx, *p)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// symbols := p.GetOpenLotSymbols()

	// for _, s := range symbols {
	// 	stdev, err := metrics.DailyStdevOfAsset(tx, s)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(s, stdev*magicNumber*100)
	// }

	fmt.Println(metrics.CalculateAssetSharpeRatio(tx, "SPY"))
	fmt.Println(metrics.CalculatePortfolioSharpeRatio(tx, *p))

}
