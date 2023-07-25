package metrics

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
			Quantity:  decimal.NewFromInt(1),
			CostBasis: decimal.NewFromInt(10),
			Trade: &domain.Trade{
				Symbol: "AAPL",
				Date:   time.Now(),
			},
		},
		{
			Quantity:  decimal.NewFromInt(1),
			CostBasis: decimal.NewFromInt(2000),
			Trade: &domain.Trade{
				Symbol: "AAPL",
				Date:   time.Now().AddDate(1, 0, 0),
			},
		},
	}
	prices := map[string]decimal.Decimal{
		"AAPL": decimal.NewFromFloat(100),
	}
	recs, err := IdentifyTLHOptions(ctx, decimal.Zero, decimal.Zero, lots, prices)
	require.NoError(t, err)
	require.Equal(t, decimal.NewFromInt(1810), recs[0].Loss)
}
