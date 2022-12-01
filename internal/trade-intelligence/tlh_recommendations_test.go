package trade_intelligence

import (
	"context"
	"hood/internal/domain"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestIdentifyTLHOptions(t *testing.T) {
	ctx := context.Background()
	lots := []domain.OpenLot{
		{
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromInt(1),
			PurchaseDate: time.Now(),
			CostBasis:    decimal.NewFromInt(10),
		},
		{
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromInt(1),
			PurchaseDate: time.Now().AddDate(1, 0, 0),
			CostBasis:    decimal.NewFromInt(2000),
		},
	}
	prices := map[string]decimal.Decimal{
		"AAPL": decimal.NewFromFloat(100),
	}
	recs, err := IdentifyTLHOptions(ctx, decimal.Zero, decimal.Zero, lots, prices)
	require.NoError(t, err)
	require.Equal(t, decimal.NewFromInt(1810), recs[0].Loss)
}
