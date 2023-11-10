package metrics

import (
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestPreviewSellOrder(t *testing.T) {
	t.Run("partial lots", func(t *testing.T) {
		trade := domain.Trade{
			Action:   model.TradeActionType_Sell,
			Symbol:   "AAPL",
			Quantity: dec(1),
			Price:    dec(100),
		}
		lots := []*domain.OpenLot{
			{
				Trade: &domain.Trade{
					Action:   model.TradeActionType_Buy,
					Symbol:   "AAPL",
					Quantity: dec(2),
					Price:    dec(100),
				},
				Quantity:  dec(2),
				CostBasis: dec(100),
			},
		}
		resp, err := PreviewSellOrder(trade, lots)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				ProcessSellOrderResult{
					NewClosedLots: []domain.ClosedLot{
						{
							OpenLot: &domain.OpenLot{
								Trade: &domain.Trade{
									Action:   model.TradeActionType_Buy,
									Symbol:   "AAPL",
									Quantity: dec(2),
									Price:    dec(100),
								},
								Quantity:  dec(1),
								CostBasis: dec(100),
							},
							SellTrade:     &trade,
							Quantity:      dec(1),
							RealizedGains: dec(0),
							GainsType:     model.GainsType_ShortTerm,
						},
					},
					OpenLots: []*domain.OpenLot{
						{
							Trade: &domain.Trade{
								Action:   model.TradeActionType_Buy,
								Symbol:   "AAPL",
								Quantity: dec(2),
								Price:    dec(100),
							},
							Quantity:  dec(1),
							CostBasis: dec(100),
						},
					},
					NewOpenLots: []domain.OpenLot{
						{
							Trade: &domain.Trade{
								Action:   model.TradeActionType_Buy,
								Symbol:   "AAPL",
								Quantity: dec(2),
								Price:    dec(100),
							},
							Quantity:  dec(1),
							CostBasis: dec(100),
						},
					},
					CashDelta: dec(100),
				},
				*resp,
			),
		)
	})
}
