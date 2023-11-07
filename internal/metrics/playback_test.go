package metrics

import (
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	. "hood/internal/domain"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestPlayback(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		times := []time.Time{
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
		}
		events := Events{
			Transfers: []domain.Transfer{{
				Amount: dec(100),
				Date:   times[0],
			}},
			Trades: []domain.Trade{
				{
					Action:   model.TradeActionType_Buy,
					Symbol:   "AAPL",
					Date:     times[1],
					Quantity: dec(1),
					Price:    dec(50),
				},
				{
					Action:   model.TradeActionType_Sell,
					Symbol:   "AAPL",
					Quantity: dec(1),
					Price:    dec(50),
					Date:     times[2],
				},
			},
		}
		dailyPortfolios, err := Playback(nil, events)
		require.NoError(t, err)

		expected := []Portfolio{
			{
				OpenLots:   map[string][]*OpenLot{},
				ClosedLots: map[string][]ClosedLot{},
				Cash:       dec(100),
				LastAction: times[0],
			},
			{
				OpenLots: map[string][]*OpenLot{
					"AAPL": {
						{
							Quantity:  dec(1),
							CostBasis: dec(50),
							Trade:     &events.Trades[0],
							Date:      times[1],
						},
					},
				},
				ClosedLots: map[string][]ClosedLot{},
				Cash:       dec(50),
				LastAction: times[1],
			},
			{
				OpenLots: map[string][]*OpenLot{},
				ClosedLots: map[string][]ClosedLot{
					"AAPL": {
						{
							RealizedGains: dec(0),
							GainsType:     model.GainsType_ShortTerm,
							Quantity:      dec(1),
							OpenLot: &domain.OpenLot{
								Quantity:  dec(0),
								CostBasis: dec(50),
								Trade:     &events.Trades[0],
								Date:      times[2],
							},
							SellTrade: &events.Trades[1],
						},
					},
				},
				Cash:       dec(100),
				LastAction: times[2],
				NewOpenLots: []domain.OpenLot{
					{
						Quantity:  dec(0),
						CostBasis: dec(50),
						Trade:     &events.Trades[0],
						Date:      times[2],
					},
				},
			},
		}

		require.Equal(
			t,
			"",
			cmp.Diff(
				expected,
				dailyPortfolios.GetPortfolios(),
			),
		)
	})

	t.Run("two sells on same day", func(t *testing.T) {
		times := []time.Time{
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		}
		events := Events{
			Trades: []Trade{
				{
					Symbol:   "AAPL",
					Action:   "BUY",
					Quantity: dec(2),
					Date:     times[0],
				},
				{
					Symbol:   "AAPL",
					Action:   "SELL",
					Quantity: dec(1),
					Date:     times[1],
				},
				{
					Symbol:   "AAPL",
					Action:   "SELL",
					Quantity: dec(1),
					Date:     times[1],
				},
			},
		}
		out, err := Playback(nil, events)
		require.NoError(t, err)

		expected := []Portfolio{
			{
				OpenLots: map[string][]*OpenLot{
					"AAPL": {
						{},
					},
				},
				ClosedLots: map[string][]ClosedLot{},
				LastAction: times[0],
			},
			{
				OpenLots: map[string][]*OpenLot{
					"AAPL": {
						{},
					},
				},
				ClosedLots: map[string][]ClosedLot{
					"AAPL": {
						{},
					},
				},
				LastAction: times[1],
				NewOpenLots: []OpenLot{
					{},
				},
			},
			{
				OpenLots: map[string][]*OpenLot{},
				ClosedLots: map[string][]ClosedLot{
					"AAPL": {
						{},
						{},
					},
				},
				LastAction: times[1],
				NewOpenLots: []OpenLot{
					{},
					{},
				},
			},
		}
		require.Equal(t, len(out.GetPortfolios()), len(expected))

		require.Equal(
			t,
			"",
			cmp.Diff(
				expected,
				out.GetPortfolios(),
				cmp.Comparer(func(x, y *OpenLot) bool {
					return x == nil && y == nil || (x != nil && y != nil)
				}),
				cmp.Comparer(func(x, y []ClosedLot) bool {
					return len(x) == len(y)
				}),
				cmp.Comparer(func(x, y []OpenLot) bool {
					return len(x) == len(y)
				}),
			),
		)
	})

}
