package metrics

import (
	"database/sql"
	"fmt"
	db "hood/internal/db/query"
	"math"
	"time"

	"github.com/shopspring/decimal"
)

func assetPriceChangeSince(tx *sql.Tx, symbol string, start time.Time) (decimal.Decimal, error) {
	adjPrices, err := db.GetAdjustedPrices(tx, []string{symbol}, start)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get adj prices: %w", err)
	}

	return (adjPrices[len(adjPrices)-1].Price.Div(adjPrices[0].Price)).Sub(decimal.NewFromInt(1)), nil
}

func MomentumFactorForAsset(tx *sql.Tx, symbol string) (decimal.Decimal, error) {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	threeMonthReturns, err := assetPriceChangeSince(tx, symbol, threeMonthsAgo)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to compute 3mo returns: %w", err)
	}
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	oneWeekReturns, err := assetPriceChangeSince(tx, symbol, oneWeekAgo)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to compute 1wk returns: %w", err)
	}
	fiveYearsAgo := time.Now().AddDate(-5, 0, 0)
	fiveYearReturns, err := assetPriceChangeSince(tx, symbol, fiveYearsAgo)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to compute 5yr returns: %w", err)
	}
	stdevF, err := DailyStdevOfAsset(tx, symbol)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to calculate stdev for momentum factor: %w", err)
	}
	stdev := decimal.NewFromFloat(stdevF * math.Sqrt(252) * 100)

	return decimal.Avg(threeMonthReturns, oneWeekReturns, fiveYearReturns).Div(stdev), nil
}
