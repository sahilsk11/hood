package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	. "hood/internal/domain"
	"hood/internal/util"
	"math"
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

	changes, err := PercentChange(prices)
	if err != nil {
		return 0, err
	}

	if err != nil {
		return 0, fmt.Errorf("failed to calculate daily percent change of %s: %w", symbol, err)
	}
	return stats.StandardDeviationSample(changes.ToStatsData())
}

func DailyStdevOfPortfolio(tx *sql.Tx, p Portfolio) (float64, error) {
	symbols := p.GetOpenLotSymbols()
	if len(symbols) == 0 {
		return 0, nil
	}
	if len(symbols) == 1 {
		return DailyStdevOfAsset(tx, symbols[0])
	}
	mappedStdev := map[string]decimal.Decimal{}
	for _, symbol := range symbols {
		stdev, err := DailyStdevOfAsset(tx, symbol)
		if err != nil {
			return 0, err
		}
		mappedStdev[symbol] = decimal.NewFromFloat(stdev)
	}

	start := time.Now().Add(-1 * stdevRange)
	covariances, err := covariances(tx, symbols, start)
	if err != nil {
		return 0, err
	}
	weights, err := assetWeights(tx, p)
	if err != nil {
		return 0, err
	}

	// idk, used a lot
	two := decimal.NewFromInt(2)

	squaredTerms := decimal.Zero
	for _, s := range symbols {
		t := (weights[s].Pow(two)).Mul(mappedStdev[s].Pow(two))
		squaredTerms = squaredTerms.Add(t)
	}

	covarianceTerms := []decimal.Decimal{}
	sort.Strings(symbols)
	for i := range symbols {
		s1 := symbols[i]
		s1Weight := weights[s1]
		s1Stdev := mappedStdev[s1]
		for j := i + 1; j < len(symbols); j++ {
			s2 := symbols[j]
			s2Weight := weights[s2]
			s2Stdev := mappedStdev[s2]
			covariance := decimal.NewFromFloat(covariances[s1+"-"+s2])
			t := two.Mul(s1Weight).Mul(s2Weight).Mul(covariance).Mul(s1Stdev).Mul(s2Stdev)

			covarianceTerms = append(covarianceTerms, t)
		}
	}
	expectedCovarianceTerms := (len(symbols) * (len(symbols) - 1)) / 2
	if len(covarianceTerms) != expectedCovarianceTerms {
		return 0, fmt.Errorf("expected %d covariance terms, calculated %d", expectedCovarianceTerms, len(covarianceTerms))
	}
	covarianceTermsSum := decimal.Zero
	for _, c := range covarianceTerms {
		covarianceTermsSum = covarianceTermsSum.Add(c)
	}

	x := math.Sqrt((squaredTerms.Add(covarianceTermsSum)).InexactFloat64())
	return x, nil
}

func assetWeights(tx *sql.Tx, p Portfolio) (map[string]decimal.Decimal, error) {
	ctx := context.Background()
	symbols := p.GetOpenLotSymbols()

	priceMap, err := db.GetLatestPrices(ctx, tx, symbols)
	if err != nil {
		return nil, err
	}
	value, err := netValue(p, priceMap)
	if err != nil {
		return nil, err
	}

	if value.Equal(decimal.Zero) {
		return nil, fmt.Errorf("portfolio has 0 net value")
	}
	weights := map[string]decimal.Decimal{}
	for symbol := range p.OpenLots {
		positionValue := p.GetQuantity(symbol).Mul(priceMap[symbol])
		weights[symbol] = positionValue.Div(value)
	}

	return weights, nil
}

func covariances(tx *sql.Tx, symbols []string, start time.Time) (map[string]float64, error) {
	if len(symbols) < 2 {
		return nil, fmt.Errorf("cannot calculate covariance of less than 2 symbols")
	}
	out := map[string]float64{}
	prices, err := db.GetAdjustedPrices(tx, symbols, start)
	if err != nil {
		return nil, err
	}

	dailyPercentChangeBySymbol, err := CalculateDailyPercentChange(prices)
	if err != nil {
		return nil, err
	}

	// necessary to produce the right
	// covariance keys
	sort.Strings(symbols)
	for i := range symbols {
		for j := i + 1; j < len(symbols); j++ {
			s1 := symbols[i]
			s2 := symbols[j]
			// https://www.investopedia.com/terms/c/covariance.asp
			c, err := stats.Covariance(
				dailyPercentChangeBySymbol[s1].ToStatsData(),
				dailyPercentChangeBySymbol[s2].ToStatsData(),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate covariance: %w", err)
			}
			key := symbols[i] + "-" + symbols[j]
			out[key] = c
		}
	}

	return out, nil
}

func setDifference(s1, s2 map[string]struct{}) []string {
	diff := make([]string, 0)
	for key := range s1 {
		if _, found := s2[key]; !found {
			diff = append(diff, key)
		}
	}
	for key := range s2 {
		if _, found := s1[key]; !found {
			diff = append(diff, key)
		}
	}
	sort.Strings(diff)
	if len(s1) != len(s2) && len(diff) == 0 {
		panic("mismatched len but could not find diff")
	}
	return diff
}

func annualExpectedReturn(tx *sql.Tx, p Portfolio) (decimal.Decimal, error) {
	weights, err := assetWeights(tx, p)
	if err != nil {
		return decimal.Zero, err
	}
	start := time.Now().Add(-1 * stdevRange)

	prices, err := db.GetAdjustedPrices(tx, p.GetOpenLotSymbols(), start)
	if err != nil {
		return decimal.Zero, err
	}
	pricesBySymbol := map[string][]model.Price{}
	for _, price := range prices {
		s := price.Symbol
		if _, ok := pricesBySymbol[s]; !ok {
			pricesBySymbol[s] = []model.Price{}
		}
		pricesBySymbol[s] = append(pricesBySymbol[s], price)
	}
	total := decimal.Zero
	for symbol, prices := range pricesBySymbol {
		sort.Slice(prices, func(i, j int) bool {
			return prices[i].Date.Before(prices[j].Date)
		})

		diffInYears := prices[len(prices)-1].Date.Sub(prices[0].Date).Hours() / (365 * 24)
		totalChange := prices[len(prices)-1].Price.Div(prices[0].Price)
		// fmt.Println(totalChange, diffInYears)
		t := decimal.NewFromFloat(math.Pow(totalChange.InexactFloat64(), (1.0 / diffInYears))).Sub(decimal.NewFromInt(1))
		total = total.Add(t.Mul(weights[symbol]))
	}
	return total, nil
}

func CalculatePortfolioSharpeRatio(tx *sql.Tx, p Portfolio) (decimal.Decimal, error) {
	magicNumber := decimal.NewFromFloat(math.Sqrt(252))
	riskFreeReturn := decimal.NewFromFloat(0.05) // approx CD return
	portfolioStdev, err := DailyStdevOfPortfolio(tx, p)
	if err != nil {
		return decimal.Zero, err
	}
	expectedReturn, err := annualExpectedReturn(tx, p)
	if err != nil {
		return decimal.Zero, err
	}

	annualStdev := decimal.NewFromFloat(portfolioStdev).Mul(magicNumber)

	util.Pprint(map[string]interface{}{
		"expectedReturn":    expectedReturn,
		"annualStdev":       annualStdev,
		"riskFeeReturnRate": riskFreeReturn,
	})

	return (expectedReturn.Sub(riskFreeReturn)).Div(annualStdev), nil
}

func CalculateAssetSharpeRatio(tx *sql.Tx, symbol string) (decimal.Decimal, error) {
	p := Portfolio{
		OpenLots: map[string][]*OpenLot{
			symbol: {
				{
					Trade: &Trade{
						Symbol:   symbol,
						Quantity: decimal.NewFromInt(1),
					},
					Quantity: decimal.NewFromInt(1),
				},
			},
		},
	}
	return CalculatePortfolioSharpeRatio(tx, p)
}

// making two design decisions here:
// 1. this layer should not have any external deps. all stateful data
// like prices and assets should be provided
// 2. it's insane to use decimal in this layer. these are all calculations
// and approximations. using decimal makes sense when dealing with specific
// amounts that we really care about. shouldn't matter for stats
//
// anyways inputs are intraday price changes (%)
func Correlation(dailyChangePricesA domain.PercentData, dailyChangePricesB domain.PercentData) (float64, error) {
	if len(dailyChangePricesA) != len(dailyChangePricesB) {
		return 0, fmt.Errorf("datasets must be same length to calculate correlation - received %d and %d", len(dailyChangePricesA), len(dailyChangePricesB))
	}

	corr, err := stats.Correlation(
		dailyChangePricesA.ToStatsData(),
		dailyChangePricesB.ToStatsData(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate correlation: %w", err)
	}

	return corr, nil
}
