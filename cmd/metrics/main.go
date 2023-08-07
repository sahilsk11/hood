package main

import (
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/portfolio"
	"hood/internal/util"
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

	out, err := portfolio.TargetAllocation(tx, map[string]decimal.Decimal{
		"TSLA": decimal.NewFromFloat(0.08),
		"NVDA": decimal.NewFromFloat(0.14),
		"AAPL": decimal.NewFromFloat(0.15),
		"MSFT": decimal.NewFromFloat(0.10),
		"UNH":  decimal.NewFromFloat(0.05),
		"BRKB": decimal.NewFromFloat(0.07),
		"GOOG": decimal.NewFromFloat(0.10),
		"META": decimal.NewFromFloat(0.05),
		"AMZN": decimal.NewFromFloat(0.14),
		"CSCO": decimal.NewFromFloat(0.02),
		"COIN": decimal.NewFromFloat(0.02),
		"C":    decimal.NewFromFloat(0.01),
		"KO":   decimal.NewFromFloat(0.05),
		"SQ":   decimal.NewFromFloat(0.02),
	})
	if err != nil {
		log.Fatal(err)
	}
	total := decimal.Zero
	for _, weight := range out {
		total = total.Add(weight)
	}
	fmt.Println(total)
	util.Pprint(out)

}
