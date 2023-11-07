package trade

import (
	"context"
	"errors"
	hood_errors "hood/internal"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func Test_tradeIngestionHandler_ProcessTdaBuyOrder(t *testing.T) {
	t.Run("try on test db", func(t *testing.T) {
		ctx := context.Background()

		dbConn, err := db.NewTest()
		require.NoError(t, err)

		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		tiService := NewTradeIngestionService()

		input := domain.Trade{
			Symbol:      "AAPL",
			Quantity:    decimal.NewFromFloat(10.5),
			Price:       decimal.NewFromFloat(100.25),
			Date:        time.Now(),
			Description: nil,
			Action:      model.TradeActionType_Buy,
			Custodian:   model.CustodianType_Tda,
		}
		id := int64(1)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, input, &id)
		require.NoError(t, err)

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, input, &id)

		require.True(t, errors.As(err, &hood_errors.ErrDuplicateTrade{}), err)
	})
}

func Test_tradeIngestionHandler_ProcessSellOrder(t *testing.T) {
	ctx := context.Background()
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	tx, err := dbConn.Begin()
	require.NoError(t, err)
	tiHandler := NewTradeIngestionService()
	db.RollbackAfterTest(t, tx)

	tr := domain.Trade{
		Symbol:    "AAPL",
		Action:    model.TradeActionType_Buy,
		Quantity:  dec(1),
		Price:     dec(100),
		Custodian: model.CustodianType_Tda,
	}

	_, _, err = tiHandler.ProcessBuyOrder(ctx, tx, tr)
	require.NoError(t, err)

	tr.Action = model.TradeActionType_Sell
	_, _, err = tiHandler.ProcessSellOrder(ctx, tx, tr)
	require.NoError(t, err)
}

func dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

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
