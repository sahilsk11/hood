package metrics

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	. "hood/internal/domain"
	"sort"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/shopspring/decimal"
)

// https://icfs.com/financial-knowledge-center/importance-standard-deviation-investment#:~:text=With%20most%20investments%2C%20including%20mutual,standard%20deviation%20would%20be%20zero.
const stdevRange = 3 * (time.Hour * 24 * 365)

func DailyStdevOfAsset(tx *sql.Tx, symbol string) (float64, error) {
	start := time.Now().Add(-1 * stdevRange)
	prices, err := db.GetAdjustedPrices(tx, []string{symbol}, start)
	if err != nil {
		return 0, err
	}

	changes, err := pricesListToMappedChanges(prices)
	if err != nil {
		return 0, err
	}
	data := decListToFloat64(changes[symbol])
	return stats.StandardDeviation(data)
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
		percentChange := ((p.Price.Sub(prevPrice)).Div(prevPrice)).Mul(decimal.NewFromInt(100))
		fmt.Printf("%s %f-%f/%f = %f\n", p.Date.Format("2006-01-02"), p.Price.InexactFloat64(), prevPrice.InexactFloat64(), prevPrice.InexactFloat64(), percentChange.InexactFloat64())
		mappedPriceLists[s] = append(changes, percentChange)
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
