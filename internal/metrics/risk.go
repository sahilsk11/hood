package metrics

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	. "hood/internal/domain"
	"math"
	"sort"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/shopspring/decimal"
)

// https://icfs.com/financial-knowledge-center/importance-standard-deviation-investment#:~:text=With%20most%20investments%2C%20including%20mutual,standard%20deviation%20would%20be%20zero.
const stdevRange = 3 * (time.Hour * 24 * 365)

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

func assetWeights(tx *sql.Tx, p Portfolio) (map[string]decimal.Decimal, error) {
	symbols := p.GetOpenLotSymbols()
	date := time.Date(2023, 7, 10, 0, 0, 0, 0, time.UTC)
	priceMap, err := getPricesHelper(tx, date, symbols)
	if err != nil {
		return nil, err
	}
	value, err := netValue(p, priceMap)
	if err != nil {
		return nil, err
	}
	weights := map[string]decimal.Decimal{}
	for symbol := range p.OpenLots {
		positionValue := p.GetQuantity(symbol).Mul(priceMap[symbol])
		weights[symbol] = positionValue.Div(value)
	}

	return weights, nil
}

// assume prices is sorted by date
func pricesListToMappedChanges(prices []model.Price) (map[string][]decimal.Decimal, error) {
	mappedPriceLists := map[string][]decimal.Decimal{}
	for i, p := range prices[1:] {
		s := p.Symbol
		prevPrice := prices[i].Price
		changes, ok := mappedPriceLists[s]
		if !ok {
			mappedPriceLists[s] = []decimal.Decimal{}
		}
		mappedPriceLists[s] = append(changes, (p.Price.Sub(prevPrice)).Div(prevPrice))
	}
	return mappedPriceLists, nil
}

func covariances(tx *sql.Tx, symbols []string, start time.Time) (map[string]float64, error) {
	out := map[string]float64{}
	prices, err := db.GetAdjustedPrices(tx, symbols, start)
	if err != nil {
		return nil, err
	}
	priceChangeMap, err := pricesListToMappedChanges(prices)
	if err != nil {
		return nil, err
	}

	sort.Strings(symbols)
	for i := range symbols {
		s1 := priceChangeMap[symbols[i]]
		for j := i; j < len(symbols); j++ {
			s2 := priceChangeMap[symbols[j]]
			if len(s1) != len(s2) {
				return nil, fmt.Errorf("inconsistent price days: %d for %s and %d for %s", len(s1), symbols[i], len(s2), symbols[j])
			}
			// https://www.investopedia.com/terms/c/covariance.asp
			c, err := stats.Covariance(
				decListToFloat64(s1),
				decListToFloat64(s2),
			)
			if err != nil {
				return nil, err
			}
			key := symbols[i] + "-" + symbols[j]
			out[key] = c
		}
	}

	return out, nil
}

func decListToFloat64(data []decimal.Decimal) stats.Float64Data {
	out := []float64{}
	for _, d := range data {
		out = append(out, d.InexactFloat64())
	}
	return out
}
