package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
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
	// util.Pprint(prices)

	changes := percentChange(prices)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate daily percent change of %s: %w", symbol, err)
	}
	data := decListToFloat64(changes)
	return stats.StandardDeviationSample(data)
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

	dailyPercentChangeBySymbol := getDailyPercentChange(prices)

	// necessary to produce the right
	// covariance keys
	sort.Strings(symbols)
	for i := range symbols {
		for j := i + 1; j < len(symbols); j++ {
			s1 := symbols[i]
			s2 := symbols[j]
			s1Data, s2Data := formatCovarianceData(s1, s2, dailyPercentChangeBySymbol)
			// https://www.investopedia.com/terms/c/covariance.asp
			c, err := stats.Covariance(
				s1Data,
				s2Data,
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

func decListToFloat64(data []decimal.Decimal) stats.Float64Data {
	out := []float64{}
	for _, d := range data {
		out = append(out, d.InexactFloat64())
	}
	return out
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

func formatCovarianceData(symbol1, symbol2 string, dailyPercentChange map[string][]decimal.Decimal) ([]float64, []float64) {
	s1Data := decListToFloat64(dailyPercentChange[symbol1])
	s2Data := decListToFloat64(dailyPercentChange[symbol2])
	// setDiff := setDifference(priceDates[s1], priceDates[s2])

	// if one asset has less values than the other,
	// use the N most recent values
	if len(s1Data) < len(s2Data) {
		// fmt.Printf("removing %d values from %s's history to match %d values from %s\n", len(s2Data)-len(s1Data), symbol2, len(s1Data), symbol1)
		s2Data = s2Data[len(s2Data)-len(s1Data):]
	} else if len(s1Data) > len(s2Data) {
		// fmt.Printf("removing %d values from %s's history to match %d values from %s\n", len(s1Data)-len(s2Data), symbol1, len(s2Data), symbol2)
		s1Data = s1Data[len(s1Data)-len(s2Data):]
	}

	return s1Data, s2Data
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

func getDailyPercentChange(prices []model.Price) map[string][]decimal.Decimal {
	pricesBySymbol := map[string][]model.Price{}

	for _, p := range prices {
		if _, ok := pricesBySymbol[p.Symbol]; !ok {
			pricesBySymbol[p.Symbol] = []model.Price{}
		}
		pricesBySymbol[p.Symbol] = append(pricesBySymbol[p.Symbol], p)
	}

	percentChangeBySymbol := map[string][]decimal.Decimal{}
	for symbol, prices := range pricesBySymbol {
		percentChangeBySymbol[symbol] = percentChange(prices)
	}

	return percentChangeBySymbol
}

func percentChange(prices []model.Price) []decimal.Decimal {
	if len(prices) < 2 {
		fmt.Println("attempted to compute percent change with less than two values")
		return []decimal.Decimal{}
	}
	sort.SliceStable(prices, func(i, j int) bool {
		return prices[i].Date.Before(prices[j].Date)
	})
	mappedPriceLists := []decimal.Decimal{}
	for i, p := range prices[1:] {
		prevPrice := prices[i].Price
		percentChange := (p.Price.Sub(prevPrice)).Div(prevPrice)
		// fmt.Printf("%s %f-%f/%f = %f\n", p.Date.Format("2006-01-02"), p.Price.InexactFloat64(), prevPrice.InexactFloat64(), prevPrice.InexactFloat64(), percentChange.InexactFloat64())
		mappedPriceLists = append(mappedPriceLists, percentChange)
	}
	return mappedPriceLists
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
