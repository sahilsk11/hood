package portfolio

import (
	"context"
	"database/sql"
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/metrics"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/shopspring/decimal"
)

type benchmark map[string]decimal.Decimal

// use a factor strategy and determine what the weight
// of each asset should be
func TargetAllocation(tx *sql.Tx, b benchmark, date time.Time) (benchmark, error) {
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

func transitionToTarget(tx *sql.Tx, currentPortfolio domain.MetricsPortfolio, target benchmark) (domain.ProposedTrades, error) {
	totalValue, err := metrics.CalculateMetricsPortfolioValue(tx, currentPortfolio, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	newQuantity, err := calculateQuantity(tx, target, totalValue)
	if err != nil {
		return nil, err
	}

	trades := []domain.ProposedTrade{}
	for symbol, position := range currentPortfolio.Positions {
		diff := newQuantity[symbol].Sub(position.Quantity)
		if !diff.Equal(decimal.Zero) {
			trades = append(trades, domain.ProposedTrade{
				Symbol:   symbol,
				Quantity: diff,
			})
		}
	}

	return trades, nil
}

func calculateQuantity(tx *sql.Tx, targetWeights benchmark, totalValue decimal.Decimal) (map[string]decimal.Decimal, error) {
	ctx := context.Background()
	symbols := []string{}
	for s := range targetWeights {
		symbols = append(symbols, s)
	}
	prices, err := db.GetLatestPrices(ctx, tx, symbols)
	if err != nil {
		return nil, err
	}
	valueBySymbol := map[string]decimal.Decimal{}
	for symbol, weight := range targetWeights {
		valueBySymbol[symbol] = totalValue.Mul(weight)
	}
	quantityBySymbol := map[string]decimal.Decimal{}
	for symbol, value := range valueBySymbol {
		quantityBySymbol[symbol] = value.Div(prices[symbol])
	}

	return quantityBySymbol, nil
}

func mpToBenchmark(tx *sql.Tx, mp domain.MetricsPortfolio) (benchmark, error) {
	out := benchmark{}
	ctx := context.Background()

	prices, err := db.GetLatestPrices(ctx, tx, mp.Symbols())
	if err != nil {
		return nil, err
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

	currentPortfolio := &initialPortfolio
	trades := []domain.Trade{}

	for start.Before(end) {
		newBenchmark, err := TargetAllocation(tx, initialBenchmark, start)
		if err != nil {
			return nil, err
		}
		proposedTrades, err := transitionToTarget(tx, *currentPortfolio, newBenchmark)
		if err != nil {
			return nil, err
		}
		err = currentPortfolio.ProcessTrades(proposedTrades)
		if err != nil {
			return nil, err
		}
		trades = append(trades, proposedTrades.ToTrades(start)...)
		start = start.AddDate(0, 0, 1)
	}

	events := Events{
		Trades: trades,
	}
	Playback(events)
}
