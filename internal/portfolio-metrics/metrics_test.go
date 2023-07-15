package metrics

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	trade "hood/internal/trade"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCalculateNetRealizedReturns(t *testing.T) {
	ctx := context.Background()
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	tiService := trade.NewTradeIngestionService()

	t.Run("net zero", func(t *testing.T) {
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		buy := newTrade(100, 1, model.TradeActionType_Buy)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, buy, 0)
		require.NoError(t, err)

		sell := newTrade(100, 1, model.TradeActionType_Sell)
		_, _, err = tiService.ProcessSellOrder(ctx, tx, sell)
		require.NoError(t, err)

		out, err := CalculateNetRealizedReturns(tx)
		require.NoError(t, err)
		require.True(t, out.Equal(decimal.Zero))
	})

	t.Run("slight gain", func(t *testing.T) {
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		buy := newTrade(100, 2, model.TradeActionType_Buy)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, buy, 0)
		require.NoError(t, err)

		sell := newTrade(110, 1, model.TradeActionType_Sell)
		_, _, err = tiService.ProcessSellOrder(ctx, tx, sell)
		require.NoError(t, err)

		sell = newTrade(130, 1, model.TradeActionType_Sell)
		_, _, err = tiService.ProcessSellOrder(ctx, tx, sell)
		require.NoError(t, err)

		out, err := CalculateNetRealizedReturns(tx)
		require.NoError(t, err)
		require.True(t, out.Equal(decimal.NewFromFloat(0.2)))
	})

	t.Run("stock split", func(t *testing.T) {
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		buy := newTrade(400, 1, model.TradeActionType_Buy)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, buy, 0)
		require.NoError(t, err)

		_, _, err = tiService.AddAssetSplit(ctx, tx, model.AssetSplit{
			Symbol:    "AAPL",
			Ratio:     4,
			Date:      time.Now(),
			CreatedAt: time.Now(),
		})
		require.NoError(t, err)

		sell := newTrade(110, 4, model.TradeActionType_Sell)
		_, _, err = tiService.ProcessSellOrder(ctx, tx, sell)
		require.NoError(t, err)

		out, err := CalculateNetRealizedReturns(tx)
		require.NoError(t, err)
		require.True(t, out.Equal(decimal.NewFromFloat(0.1)))
	})

}

func TestCalculateNetUnrealizedReturns(t *testing.T) {
	ctx := context.Background()
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	tiService := trade.NewTradeIngestionService()

	t.Run("net zero", func(t *testing.T) {
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		buy := newTrade(100, 1, model.TradeActionType_Buy)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, buy, 0)
		require.NoError(t, err)

		_, err = db.AddPrices(ctx, tx, []model.Price{
			{
				Symbol:    "AAPL",
				Price:     decimal.NewFromFloat(100),
				UpdatedAt: time.Now(),
			},
		})
		require.NoError(t, err)

		out, err := CalculateNetUnrealizedReturns(tx)
		require.NoError(t, err)
		require.True(t, out.Equal(decimal.Zero))
	})

	t.Run("slight gain", func(t *testing.T) {
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		buy := newTrade(100, 1, model.TradeActionType_Buy)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, buy, 0)
		require.NoError(t, err)

		_, err = db.AddPrices(ctx, tx, []model.Price{
			{
				Symbol:    "AAPL",
				Price:     decimal.NewFromFloat(120),
				UpdatedAt: time.Now(),
			},
		})
		require.NoError(t, err)

		out, err := CalculateNetUnrealizedReturns(tx)
		require.NoError(t, err)
		require.True(t, out.Equal(decimal.NewFromFloat(0.2)))
	})
}

func Test_CalculateNetReturns(t *testing.T) {
	ctx := context.Background()
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	tiService := trade.NewTradeIngestionService()

	t.Run("net zero", func(t *testing.T) {
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		buy := newTrade(100, 1, model.TradeActionType_Buy)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, buy, 0)
		require.NoError(t, err)
		sell := newTrade(120, 1, model.TradeActionType_Sell)
		_, _, err = tiService.ProcessSellOrder(ctx, tx, sell)
		require.NoError(t, err)

		buy = newTrade(100, 1, model.TradeActionType_Buy)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, buy, 1)
		require.NoError(t, err)
		_, err = db.AddPrices(ctx, tx, []model.Price{
			{
				Symbol:    "AAPL",
				Price:     decimal.NewFromFloat(80),
				UpdatedAt: time.Now(),
			},
		})
		require.NoError(t, err)

		out, err := CalculateNetReturns(tx)
		require.NoError(t, err)

		require.True(t, out.Equal(decimal.Zero), out)
	})
}

func newTrade(price, quantity float64, action model.TradeActionType) model.Trade {
	return model.Trade{
		Symbol:      "AAPL",
		Action:      action,
		Quantity:    decimal.NewFromFloat(quantity),
		CostBasis:   decimal.NewFromFloat(price),
		Date:        time.Now(),
		Description: nil,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		Custodian:   model.CustodianType_Tda,
	}
}
