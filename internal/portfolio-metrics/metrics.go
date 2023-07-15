package metrics

import (
	"database/sql"
	"fmt"
	db "hood/internal/db/query"

	"github.com/shopspring/decimal"
)

func CalculateNetReturns(tx *sql.Tx) (decimal.Decimal, error) {
	totalRealizedGains, err := db.GetTotalRealizedGains(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total realized gains: %w", err)
	}
	totalRealizedCostBasis, err := db.GetTotalRealizedCostBasis(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total realized cost basis: %w", err)
	}
	totalUnrealizedGains, err := db.GetTotalUnrealizedGains(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total unrealized gains: %w", err)
	}
	totalUnrealizedCostBasis, err := db.GetTotalUnrealizedCostBasis(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total unrealized cost basis: %w", err)
	}

	totalGains := totalUnrealizedGains.Add(totalRealizedGains)
	totalCostBasis := totalUnrealizedCostBasis.Add(totalRealizedCostBasis)
	if totalCostBasis.Equal(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("received 0 total cost basis: %w", err)
	}
	return totalGains.Div(totalCostBasis), nil
}

func CalculateNetRealizedReturns(tx *sql.Tx) (decimal.Decimal, error) {
	totalRealizedGains, err := db.GetTotalRealizedGains(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total realized gains: %w", err)
	}
	totalRealizedCostBasis, err := db.GetTotalRealizedCostBasis(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total realized cost basis: %w", err)
	}

	if totalRealizedCostBasis.Equal(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("received 0 total cost basis: %w", err)
	}

	return totalRealizedGains.Div(totalRealizedCostBasis), nil
}

func CalculateNetUnrealizedReturns(tx *sql.Tx) (decimal.Decimal, error) {
	totalUnrealizedGains, err := db.GetTotalUnrealizedGains(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total unrealized gains: %w", err)
	}
	totalUnrealizedCostBasis, err := db.GetTotalUnrealizedCostBasis(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total unrealized cost basis: %w", err)
	}

	if totalUnrealizedCostBasis.Equal(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("received 0 total cost basis: %w", err)
	}

	return totalUnrealizedGains.Div(totalUnrealizedCostBasis), nil
}
