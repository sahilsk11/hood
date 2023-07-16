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
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}

	for _, result := range results {
		priceMap[result.Symbol] = result.Price
		symbolSet[result.Symbol] = true
	}

	// ensure result map has all requested symbols
	for _, s := range symbols {
		if !symbolSet[s] {
			return nil, fmt.Errorf("symbol %s does not have price updated on %s", s, date.Format("2006-01-02"))
		}
	}

	return priceMap, nil
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
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
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
