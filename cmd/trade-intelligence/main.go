package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/domain"
	trade_intelligence "hood/internal/trade-intelligence"
	"log"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func main() {
	ctx := context.Background()

	connStr := "postgresql://postgres:postgres@localhost:5438/postgres?sslmode=disable"
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := dbConn.Begin()
	if err != nil {
		log.Fatal(err)
	}

	lots, err := db.GetVwOpenLotPosition(ctx, tx)
	if err != nil {
		log.Fatal(err)
	}

	// ok to have duplicates because we use SQL IN
	symbols := []string{}
	domainLots := []domain.OpenLot{}
	for _, lot := range lots {
		domainLots = append(domainLots, domain.OpenLotFromVwOpenLotPosition(lot))
		symbols = append(symbols, *lot.Symbol)
	}

	priceMap, err := db.GetPrices(ctx, tx, symbols)
	if err != nil {
		log.Fatal(err)
	}

	results, err := trade_intelligence.IdentifyTLHOptions(
		ctx,
		decimal.NewFromInt(5),
		decimal.NewFromInt(1),
		domainLots,
		priceMap,
	)
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
