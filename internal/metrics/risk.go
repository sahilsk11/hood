package metrics

import (
	"database/sql"
	"fmt"
	db "hood/internal/db/query"
	"math"

	"github.com/montanaflynn/stats"
	"github.com/shopspring/decimal"
)

// functions to generate risk metrics
func StandardDeviation(
	tx *sql.Tx,
	dailyChange map[string]decimal.Decimal,
) (float64, error) {
	holidays, err := tradingHolidays(tx)
	if err != nil {
		return 0, fmt.Errorf("failed to get trading holidays")
	}

	values := stats.Float64Data{}
	for k, v := range dailyChange {
		_, isTradingHoliday := holidays[k]
		if !isTradingHoliday {
			values = append(values, v.InexactFloat64())
		}
	}

	stdev, err := stats.StandardDeviation(values)
	if err != nil {
		return 0, fmt.Errorf("failed to compute stdev: %w", err)
	}

	magicNumber := math.Pow(252, 0.5)

	return (stdev * magicNumber) * 100, nil
}

func StdevOfAsset(tx *sql.Tx, symbol string) (float64, error) {
	prices, err := db.GetPricesChanges(tx, symbol)
	if err != nil {
		return 0, err
	}
	fmt.Println(prices["2021-01-01"])
	return StandardDeviation(tx, prices)
}

func tradingHolidays(tx *sql.Tx) (map[string]struct{}, error) {
	holidays := map[string]struct{}{}
	priceDates, err := db.DistinctPriceDays(tx)
	if err != nil {
		return nil, err
	}

	if len(priceDates) == 0 {
		fmt.Println("no holidays found")
		return holidays, nil
	}
	priceDatesSet := map[string]struct{}{}
	for _, d := range priceDates {
		priceDatesSet[d.Format(layout)] = struct{}{}
	}
	start := priceDates[0]
	end := priceDates[len(priceDates)-1]

	for start.Before(end) || start.Equal(end) {
		dateStr := start.Format(layout)
		if _, ok := priceDatesSet[dateStr]; !ok {
			holidays[dateStr] = struct{}{}
		}
		start = start.AddDate(0, 0, 1)
	}

	return holidays, nil
}
