package service

import (
	"database/sql"
	db "hood/internal/db/query"
	"hood/internal/metrics"
	"sort"

	"time"
)

// https://icfs.com/financial-knowledge-center/importance-standard-deviation-investment#:~:text=With%20most%20investments%2C%20including%20mutual,standard%20deviation%20would%20be%20zero.
const stdevRange = 3 * (time.Hour * 24 * 365)

type AssetCorrelation struct {
	AssetOne    string
	AssetTwo    string
	Correlation float64
}

func PortfolioCorrelation(tx *sql.Tx, symbols []string) ([]AssetCorrelation, error) {
	start := time.Now().Add(-1 * stdevRange)
	prices, err := db.GetAdjustedPrices(tx, symbols, start)
	if err != nil {
		return nil, err
	}

	dailyPercentChanges, err := metrics.CalculateDailyPercentChange(prices)
	if err != nil {
		return nil, err
	}

	sort.Strings(symbols)

	out := []AssetCorrelation{}
	for i, s1 := range symbols {
		for j := i + 1; j < len(symbols); j++ {
			s2 := symbols[j]
			corr, err := metrics.Correlation(dailyPercentChanges[s1], dailyPercentChanges[s2])
			if err != nil {
				return nil, err
			}
			out = append(out, AssetCorrelation{
				AssetOne:    s1,
				AssetTwo:    s2,
				Correlation: corr,
			})
		}
	}

	return out, nil
}
