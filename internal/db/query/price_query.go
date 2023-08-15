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

func AddPrices(tx *sql.Tx, prices []model.Price) ([]model.Price, error) {
	out := []model.Price{}
	batchSize := 500
	offset := 0

	for offset < len(prices) {
		end := offset + batchSize
		if len(prices) < end {
			end = len(prices)
		}
		batch := prices[offset:end]

		for i := range prices {
			prices[i].UpdatedAt = time.Now().UTC()
		}
		t := Price
		stmt := t.INSERT(t.MutableColumns).
			MODELS(batch).
			RETURNING(t.AllColumns)

		result := []model.Price{}
		err := stmt.Query(tx, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to insert prices: %w", err)
		}
		out = append(out, result...)
		offset += len(batch)
	}
	return out, nil
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

func GetPricesHelper(tx *sql.Tx, date time.Time, symbols []string) (map[string]decimal.Decimal, error) {
	if len(symbols) == 0 {
		return map[string]decimal.Decimal{}, nil
	}
	priceMap, err := GetPricesOnDate(tx, date, symbols)
	if err != nil {
		e := err
		tries := 5
		for tries > 0 && e != nil {
			date = date.AddDate(0, 0, -1)
			priceMap, e = GetPricesOnDate(tx, date, symbols)
			tries -= 1
		}
		if e != nil {
			return nil, fmt.Errorf("failed to get prices: %w", err)
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
	if len(symbols) == 0 {
		return nil, fmt.Errorf("cannot query prices for 0 given symbols")
	}
	priceMap := map[string]decimal.Decimal{}
	symbolSet := map[string]bool{}

	postgresStr := []Expression{}
	for _, s := range symbols {
		symbolSet[s] = false
		postgresStr = append(postgresStr, String(s))
	}
	t := view.LatestPrice
	query := t.SELECT(t.AllColumns).
		WHERE(t.Symbol.IN(postgresStr...))

	results := []model.LatestPrice{}
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
	// apply asset splits
	for i, r := range result {
		for _, split := range assetSplits {
			if r.Symbol == split.Symbol && r.Date.Before(split.Date) {
				result[i].Price = result[i].Price.Div(decimal.NewFromInt32(split.Ratio))
			}
		}
	}

	return result, nil
}

func AddDjMetrics(tx *sql.Tx, metrics []model.DataJockeyAssetMetrics) error {
	query := DataJockeyAssetMetrics.INSERT(DataJockeyAssetMetrics.MutableColumns).MODELS(metrics)

	_, err := query.Exec(tx)
	if err != nil {
		return err
	}
	return nil
}

func GetDjMetrics(tx *sql.Tx, symbol string) (*model.DataJockeyAssetMetrics, error) {
	query := DataJockeyAssetMetrics.
		SELECT(DataJockeyAssetMetrics.AllColumns).
		WHERE(DataJockeyAssetMetrics.Symbol.EQ(String(symbol))).
		ORDER_BY(DataJockeyAssetMetrics.CreatedAt).
		LIMIT(1)

	out := &model.DataJockeyAssetMetrics{}
	err := query.Query(tx, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
