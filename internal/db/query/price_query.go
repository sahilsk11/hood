package db

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/db/models/postgres/public/view"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/shopspring/decimal"
)

func AddPrices(ctx context.Context, tx *sql.Tx, prices []model.Price) ([]model.Price, error) {
	t := Price
	stmt := t.INSERT(t.MutableColumns).
		MODELS(prices).
		RETURNING(t.AllColumns)

	result := []model.Price{}
	err := stmt.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to insert prices: %w", err)
	}

	return result, nil
}

func GetPricesOnDate(tx *sql.Tx, date time.Time, symbols []string) (map[string]decimal.Decimal, error) {
	priceMap := map[string]decimal.Decimal{}
	symbolSet := map[string]bool{}

	postgresStr := []Expression{}
	for _, s := range symbols {
		symbolSet[s] = false
		postgresStr = append(postgresStr, String(s))
	}

	whereExp := []BoolExpression{
		Price.Date.EQ(DateT(date)),
	}
	if len(symbols) > 0 {
		whereExp = append(whereExp,
			Price.Symbol.IN(postgresStr...),
		)
	}

	query := Price.SELECT(Price.AllColumns).
		WHERE(AND(whereExp...))

	results := []model.Price{}
	err := query.Query(tx, &results)
	if err != nil {
		fmt.Println(query.DebugSql())
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	for _, result := range results {
		priceMap[result.Symbol] = result.Price
		symbolSet[result.Symbol] = true
	}

	// ensure result map has all requested symbols
	for _, s := range symbols {
		if !symbolSet[s] && s != "AMAG" && s != "ETH" && s != "BTC" && s != "DOGE" {
			return nil, fmt.Errorf("symbol %s does not have price updated on %s", s, date.Format("2006-01-02"))
		}
	}

	return priceMap, nil
}

func GetPricesChanges(tx *sql.Tx, symbol string) (map[string]decimal.Decimal, error) {
	query := Price.SELECT(Price.Date, Price.Price).
		WHERE(AND(
			Price.Symbol.EQ(String(symbol)),
			Price.Date.GT(Date(2023, 1, 1)),
		)).
		ORDER_BY(Price.Date.ASC())

	result := []model.Price{}
	err := query.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}

	out := map[string]decimal.Decimal{}
	for i, r := range result[1:] {
		prevPrice := result[i].Price
		out[r.Date.Format("2006-01-02")] = (r.Price.Sub(prevPrice)).Div(prevPrice)
	}

	return out, nil

}

func GetLatestPrices(ctx context.Context, tx *sql.Tx, symbols []string) (map[string]decimal.Decimal, error) {
	priceMap := map[string]decimal.Decimal{}
	symbolSet := map[string]bool{}

	postgresStr := []Expression{}
	for _, s := range symbols {
		symbolSet[s] = false
		postgresStr = append(postgresStr, String(s))
	}
	t := view.VwLatestPrice
	query := t.SELECT(t.AllColumns).
		WHERE(t.Symbol.IN(postgresStr...))

	results := []model.VwLatestPrice{}
	err := query.Query(tx, &results)
	if err != nil {
		fmt.Println(query.DebugSql())
		return nil, fmt.Errorf("failed to execute latest price query: %w", err)
	}

	for _, result := range results {
		priceMap[*result.Symbol] = *result.Price
		symbolSet[*result.Symbol] = true
	}

	// ensure result map has all requested symbols
	for _, s := range symbols {
		if !symbolSet[s] {
			return nil, fmt.Errorf("symbol %s does not have latest price updated", s)
		}
	}

	return priceMap, nil
}

func DistinctPriceDays(tx *sql.Tx) ([]time.Time, error) {
	var dates []time.Time
	query := SELECT(Price.Date).
		FROM(Price).
		DISTINCT().
		ORDER_BY(Price.Date.ASC())
	err := query.Query(tx, &dates)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch distinct dates: %w", err)
	}
	return dates, nil
}

func GetAssetSplits(tx *sql.Tx, symbols []string) ([]model.AssetSplit, error) {
	symbolExpression := symbolExpression(symbols)
	query := AssetSplit.SELECT(AssetSplit.AllColumns).
		WHERE(AssetSplit.Symbol.IN(symbolExpression...))

	result := []model.AssetSplit{}
	err := query.Query(tx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	return result, nil
}

func symbolExpression(symbols []string) []Expression {
	symbolExpression := []Expression{}
	for _, s := range symbols {
		symbolExpression = append(symbolExpression, String(s))
	}
	return symbolExpression
}

// prices adjusted for asset splits
func GetAdjustedPrices(tx *sql.Tx, symbols []string, start time.Time) ([]model.Price, error) {
	symbolExpression := symbolExpression(symbols)
	query := Price.SELECT(Price.AllColumns).
		WHERE(AND(
			Price.Symbol.IN(symbolExpression...),
			Price.Date.GT_EQ(Date(start.Date())),
		)).
		ORDER_BY(Price.Date.ASC())

	result := []model.Price{}
	err := query.Query(tx, &result)
	if err != nil {
		fmt.Println(query.DebugSql())
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}

	assetSplits, err := GetAssetSplits(tx, symbols)
	if err != nil {
		return nil, err
	}

	for i, r := range result {
		for _, split := range assetSplits {
			if r.Symbol == split.Symbol && r.Date.Before(split.Date) {
				result[i].Price = result[i].Price.Mul(decimal.NewFromInt32(split.Ratio))
			}
		}
	}

	return result, nil
}
