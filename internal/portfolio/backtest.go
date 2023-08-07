package portfolio

import (
	"database/sql"
	"fmt"
	"hood/internal/metrics"
	"hood/internal/util"

	"github.com/montanaflynn/stats"
	"github.com/shopspring/decimal"
)

type benchmark map[string]decimal.Decimal

func TargetAllocation(tx *sql.Tx, b benchmark) (benchmark, error) {
	symbols := []string{}
	momentumFactorBySymbol := map[string]decimal.Decimal{}
	momentumFactorValues := []decimal.Decimal{}
	for k := range b {
		symbols = append(symbols, k)
		mFactor, err := metrics.MomentumFactorForAsset(tx, k)
		if err != nil {
			return nil, err
		}
		momentumFactorValues = append(momentumFactorValues, mFactor)
		momentumFactorBySymbol[k] = mFactor
	}
	util.Pprint(momentumFactorBySymbol)
	dataset := []float64{}
	for _, m := range momentumFactorValues {
		dataset = append(dataset, m.InexactFloat64())
	}
	mFactorsMean := decimal.Avg(momentumFactorValues[0], momentumFactorValues[1:]...)
	mFactorsStdevF, err := stats.StandardDeviationSample(dataset)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate stdev of momentum factors: %w", err)
	}
	mFactorsStdev := decimal.NewFromFloat(mFactorsStdevF)
	fmt.Println(mFactorsMean, mFactorsStdev)
	zScoreBySymbol := map[string]decimal.Decimal{}
	maxScaleFactor := decimal.NewFromInt(1)
	for symbol, mFactor := range momentumFactorBySymbol {
		zScore := (mFactor.Sub(mFactorsMean)).Div(mFactorsStdev)
		maxB := (decimal.NewFromInt(1).Sub(b[symbol])).Div(zScore)
		if zScore.LessThan(decimal.Zero) {
			maxB = b[symbol].Div(zScore).Neg()
		}
		if maxB.LessThan(maxScaleFactor) {
			maxScaleFactor = maxB
		}
		zScoreBySymbol[symbol] = zScore
	}
	util.Pprint(zScoreBySymbol)

	// set intensity of factor weight
	scaleFactor := maxScaleFactor.Mul(decimal.NewFromFloat(0.98))

	out := benchmark{}
	for symbol, originalWeight := range b {
		out[symbol] = originalWeight.Add(scaleFactor.Mul(zScoreBySymbol[symbol]))
	}

	return out, nil
}
