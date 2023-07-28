package metrics

import (
	"fmt"
	"hood/internal/util"
	"sort"

	"github.com/shopspring/decimal"
)

func DailyAggregateTwr(dailyPortfolioValues map[string]decimal.Decimal, transfers map[string]decimal.Decimal) (map[string]decimal.Decimal, error) {
	dailyTwr, err := TimeWeightedReturns(dailyPortfolioValues, transfers)
	if err != nil {
		return nil, err
	}
	aggregateTwr := decimal.NewFromInt(1)
	out := map[string]decimal.Decimal{}
	dates := util.SortedMapKeys(dailyTwr)
	for _, d := range dates {
		aggregateTwr = aggregateTwr.Mul(dailyTwr[d])
		out[d] = aggregateTwr
	}
	return out, nil
}

// calculates twr product for each day
// in daily portfolio values. should be
// equivalent to "daily returns"
func TimeWeightedReturns(
	dailyPortfolioValues map[string]decimal.Decimal,
	transfers map[string]decimal.Decimal,
) (map[string]decimal.Decimal, error) {
	if len(dailyPortfolioValues) < 2 {
		return nil, fmt.Errorf("at least two daily portfolios required to compute TWR")
	}
	dateKeys := []string{}
	for dateStr := range dailyPortfolioValues {
		dateKeys = append(dateKeys, dateStr)
	}
	sort.Strings(dateKeys)

	out := map[string]decimal.Decimal{}

	for i := 1; i < len(dateKeys); i++ {
		prevValue := dailyPortfolioValues[dateKeys[i-1]]
		currentValue := dailyPortfolioValues[dateKeys[i]]
		netTransfers, _ := transfers[dateKeys[i]]

		hp := hp(
			prevValue,
			currentValue,
			netTransfers,
		)
		out[dateKeys[i]] = hp
	}

	return out, nil
}

// https://www.investopedia.com/terms/t/time-weightedror.asp
func hp(start, end, cashFlow decimal.Decimal) decimal.Decimal {
	initialPlusTransfers := start.Add(cashFlow)
	numerator := end
	denominator := initialPlusTransfers

	quotient := numerator.Div(denominator)
	hp := quotient

	util.Pprint(map[string]decimal.Decimal{
		"hp":          hp,
		"numerator":   numerator,
		"denominator": denominator,
		"start":       start,
		"end":         end,
		"cashFlows":   cashFlow,
	})

	return hp
}
