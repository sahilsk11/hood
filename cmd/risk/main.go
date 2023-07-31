package main

import (
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/metrics"
	"hood/internal/portfolio"
	"log"
	"math"
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

	p, err := portfolio.GetAggregatePortfolio(tx)
	if err != nil {
		log.Fatal(err)
	}

	// portfolioStdev, err := metrics.DailyStdevOfPortfolio(tx, *p)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	magicNumber := math.Sqrt(252)

	symbols := p.GetOpenLotSymbols()

	for _, s := range symbols {
		stdev, err := metrics.DailyStdevOfAsset(tx, s)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(s, stdev*magicNumber*100)
	}

	stdev, err := metrics.DailyStdevOfAsset(tx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(stdev * magicNumber * 100)
}
