package db

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestGetOpenLots(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		ctx := context.Background()
		dbConn, err := NewTest()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		RollbackAfterTest(t, tx)

		trades := []domain.Trade{
			{
				Symbol:           "AAPL",
				Quantity:         dec(10),
				Price:            dec(100),
				Action:           model.TradeActionType_Buy,
				TradingAccountID: uuid.New(),
			},
		}
		trades, err = AddTrades(ctx, tx, trades)
		require.NoError(t, err)
		openLots := []domain.OpenLot{
			{
				Trade:     &trades[0],
				CostBasis: dec(100),
				Quantity:  dec(10),
			},
		}
		openLots, err = AddOpenLots(ctx, tx, openLots)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				[]domain.OpenLot{
					{
						CostBasis: dec(100),
						Quantity:  dec(10),
						Trade:     &trades[0],
					},
				},
				openLots,
				cmpopts.IgnoreFields(domain.Trade{}, "TradeID"),
				cmpopts.IgnoreFields(domain.OpenLot{}, "OpenLotID"),
			),
		)
	})
}

func dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}
