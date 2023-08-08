package portfolio

import (
	"context"
	"database/sql"
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/metrics"
	"hood/internal/util"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/shopspring/decimal"
)

type benchmark map[string]decimal.Decimal

// use a factor strategy and determine what the weight
// of each asset should be
func TargetAllocation(tx *sql.Tx, b benchmark, date time.Time) (benchmark, error) {
	if len(b) < 2 {
		return nil, fmt.Errorf("cannot calculate allocation with < 2 assets")
	}
	symbols := []string{}
	momentumFactorBySymbol := map[string]decimal.Decimal{}
	momentumFactorValues := []decimal.Decimal{}
	for k := range b {
		symbols = append(symbols, k)
		mFactor, err := metrics.MomentumFactorForAsset(tx, k, date)
		if err != nil {
			return nil, err
		}
		momentumFactorValues = append(momentumFactorValues, mFactor)
		momentumFactorBySymbol[k] = mFactor
	}

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

	// set intensity of factor weight
	scaleFactor := maxScaleFactor.Mul(decimal.NewFromFloat(0.98))

	out := benchmark{}
	for symbol, originalWeight := range b {
		out[symbol] = originalWeight.Add(scaleFactor.Mul(zScoreBySymbol[symbol]))
	}

	return out, nil
}

func transitionToTarget(tx *sql.Tx, currentPortfolio domain.MetricsPortfolio, target benchmark, date time.Time) (domain.ProposedTrades, error) {
	util.Pprint(target)
	totalValue, err := metrics.CalculateMetricsPortfolioValue(tx, currentPortfolio, date)
	if err != nil {
		return nil, err
	}

	prices, err := db.GetPricesHelper(tx, date, currentPortfolio.Symbols())
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	newQuantity, err := calculateQuantity(prices, target, totalValue)
	if err != nil {
		return nil, err
	}
	// if quantities are too low, skip trade
	for _, v := range newQuantity {
		if v.Abs().LessThan(decimal.NewFromFloat(0.0001)) {
			return []domain.ProposedTrade{}, nil
		}
	}

	trades := []domain.ProposedTrade{}
	for symbol, position := range currentPortfolio.Positions {
		diff := newQuantity[symbol].Sub(position.Quantity)
		if !diff.Equal(decimal.Zero) {
			trades = append(trades, domain.ProposedTrade{
				Symbol:        symbol,
				Quantity:      diff,
				ExpectedPrice: prices[symbol],
			})
		}
	}

	return trades, nil
}

func calculateQuantity(priceMap map[string]decimal.Decimal, targetWeights benchmark, totalValue decimal.Decimal) (map[string]decimal.Decimal, error) {
	symbols := []string{}
	for s := range targetWeights {
		symbols = append(symbols, s)
	}

	valueBySymbol := map[string]decimal.Decimal{}
	for symbol, weight := range targetWeights {
		valueBySymbol[symbol] = totalValue.Mul(weight)
	}
	quantityBySymbol := map[string]decimal.Decimal{}
	for symbol, value := range valueBySymbol {
		quantityBySymbol[symbol] = value.Div(priceMap[symbol])
	}

	return quantityBySymbol, nil
}

func mpToBenchmark(tx *sql.Tx, mp domain.MetricsPortfolio) (benchmark, error) {
	if len(mp.Symbols()) == 0 {
		return nil, fmt.Errorf("cannot convert empty metrics portfolio to benchmark")
	}
	out := benchmark{}
	ctx := context.Background()

	prices, err := db.GetLatestPrices(ctx, tx, mp.Symbols())
	if err != nil {
		return nil, fmt.Errorf("failed to get prices for symbols %v: %w", mp.Symbols(), err)
	}
	valueBySymbol := map[string]decimal.Decimal{}
	for symbol, position := range mp.Positions {
		valueBySymbol[symbol] = position.Quantity.Mul(prices[symbol])
	}
	totalValue, err := metrics.CalculateMetricsPortfolioValue(tx, mp, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	for symbol := range mp.Positions {
		out[symbol] = valueBySymbol[symbol].Div(totalValue)
	}

	return out, nil
}

func Backtest(
	tx *sql.Tx,
	initialPortfolio domain.MetricsPortfolio,
	start time.Time,
	end time.Time,
) (*domain.HistoricPortfolio, error) {
	initialBenchmark, err := mpToBenchmark(tx, initialPortfolio)
	if err != nil {
		return nil, err
	}

	currentPortfolio := initialPortfolio.DeepCopy()
	trades := []domain.Trade{}
	current := start

	// i := 0
	for current.Before(end) {
		// i++
		// if i%50 == 0 {
		// 	fmt.Println(start, end)
		// 	fmt.Println(*currentPortfolio.Positions["AAPL"], *currentPortfolio.Positions["MSFT"])
		// }
		newBenchmark, err := TargetAllocation(tx, initialBenchmark, current)
		if err != nil {
			return nil, err
		}
		proposedTrades, err := transitionToTarget(tx, *currentPortfolio, newBenchmark, current)
		if err != nil {
			return nil, err
		}
		err = currentPortfolio.ProcessTrades(proposedTrades)
		if err != nil {
			return nil, err
		}
		trades = append(trades, proposedTrades.ToTrades(current)...)
		current = current.AddDate(0, 0, 30)
	}

	events := Events{
		Trades: trades,
	}
	out, err := Playback(initialPortfolio.NewPortfolio(nil, start), events)
	if err != nil {
		return nil, err
	}

	return out, nil
}
