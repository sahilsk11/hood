package portfolio

import (
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/domain"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPlayback(t *testing.T) {
	times := []time.Time{
		time.Now().Add(-3 * time.Second),
		time.Now().Add(-2 * time.Second),
		time.Now().Add(-1 * time.Second),
	}
	trades := []Trade{
		{
			Symbol:   "AAPL",
			Quantity: dec(10),
			Price:    dec(100),
			Action:   model.TradeActionType_Buy,
			Date:     times[1],
		},
		{
			Symbol:   "AAPL",
			Quantity: dec(10),
			Price:    dec(100),
			Action:   model.TradeActionType_Sell,
			Date:     times[2],
		},
	}
	transfers := []Transfer{{Amount: dec(1000), Date: times[0]}}
	out, err := Playback(trades, nil, transfers)
	require.NoError(t, err)
	require.Equal(
		t,
		"",
		cmp.Diff(
			Portfolio{
				OpenLots: map[string][]OpenLot{},
				ClosedLots: map[string][]ClosedLot{
					"AAPL": {
						{
							OpenLot: &OpenLot{
								Trade:     &trades[0],
								Quantity:  dec(0),
								CostBasis: dec(100),
								Date:      times[1],
							},
							Quantity:      dec(10),
							GainsType:     model.GainsType_ShortTerm,
							SellTrade:     &trades[1],
							RealizedGains: dec(0),
						},
					},
				},
				Date: times[2],
				Cash: dec(1000),
			},
			*out,
			cmpopts.IgnoreFields(OpenLot{}, "LotID"),
		),
	)
}

func dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}
