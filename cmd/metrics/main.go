package main

import (
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/metrics"
	"hood/internal/portfolio"
	"log"

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

	start := domain.MetricsPortfolio{
		Positions: map[string]*domain.Position{
			"NVDA": {
				Quantity: decimal.NewFromInt(25),
			},
			"CSCO": {
				Quantity: decimal.NewFromInt(250),
			},
			"AAPL": {
				Quantity: decimal.NewFromInt(60),
			},
		},
	}

	hp1 := domain.NewHistoricPortfolio(
		[]domain.Portfolio{
			*start.NewPortfolio(nil, time.Now().AddDate(-1, 0, 0)),
		},
	)

	hp, err := portfolio.Backtest(tx, start, time.Now().AddDate(-1, 0, 0), time.Now())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("finished backtest")

	r1, err := metrics.CalculateTotalReturn(tx, *hp)
	if err != nil {
		log.Fatal(err)
	}
	r2, err := metrics.CalculateTotalReturn(tx, *hp1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r2.InexactFloat64(), r1.InexactFloat64())
}
