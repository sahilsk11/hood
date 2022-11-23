package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	trade_intelligence "hood/internal/trade-intelligence"
	"log"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func main() {
	connStr := "postgresql://postgres:postgres@localhost:5438/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.WithValue(
		context.Background(),
		"tx",
		tx,
	)

	results, err := trade_intelligence.IdentifyTLHOptions(ctx)
	if err != nil {
		log.Fatal(err)
	}
	total := decimal.Zero
	for _, r := range results {
		if r.BreakevenPriceChange.GreaterThan(decimal.NewFromInt(0)) {
			total = total.Add(r.Loss)

		}
		// fmt.Printf("%s\t\t\t\t%f\t\t%f\t\t\t%f\n", r.Symbol, r.SellQuantity.InexactFloat64(), r.Loss.InexactFloat64(), r.BreakevenPriceChange.InexactFloat64())
	}
	b, _ := json.Marshal(results)
	fmt.Println(string(b))
}
