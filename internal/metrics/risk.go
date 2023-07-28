package metrics

import (
	"database/sql"
	"fmt"
	db "hood/internal/db/query"

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
		return 0, err
	}
	values := stats.Float64Data{}
	for k, v := range dailyChange {
		if _, ok := holidays[k]; !ok {
			fmt.Println(v)
			values = append(values, v.InexactFloat64())
		}
	}

	return stats.StandardDeviation(values)
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
	for start.Before(end) {
		dateStr := start.Format(layout)
		if _, ok := priceDatesSet[dateStr]; !ok {
			holidays[dateStr] = struct{}{}
		}
		start = start.AddDate(0, 0, 1)
	}
	return holidays, nil
}
