package main

import (
	db "hood/internal/db/query"
	"hood/internal/prices"
	"log"

	"github.com/piquette/finance-go/equity"
)

// This file lists several usage examples of this library
// and can be used to verify behavior.
func main() {

	equity.GetHistoricalQuote()

	quote.
		dbConn, err := db.New()
	if err != nil {
		log.Fatal(err)
	}
	err = prices.UpdateFromYahoo(dbConn)
	if err != nil {
		log.Fatal(err)
	}
}
