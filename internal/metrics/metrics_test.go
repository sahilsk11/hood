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

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade.ProcessTdaBuyOrderInput{
			Symbol:           "AAPL",
			TdaTransactionID: 0,
			Quantity:         decimal.NewFromFloat(1),
			CostBasis:        decimal.NewFromFloat(100),
			Date:             time.Now(),
			Description:      nil,
		})
		require.NoError(t, err)

		_, _, err = tiService.ProcessSellOrder(ctx, tx, trade.ProcessSellOrderInput{
			Symbol:    "AAPL",
			Quantity:  decimal.NewFromFloat(1),
			CostBasis: decimal.NewFromFloat(100),
			Date:      time.Now(),
			Custodian: model.CustodianType_Tda,
		})
		require.NoError(t, err)

		out, err := CalculateNetRealizedReturns(tx)
		require.NoError(t, err)
		require.True(t, out.Equal(decimal.Zero))
	})

	t.Run("slight gain", func(t *testing.T) {
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade.ProcessTdaBuyOrderInput{
			Symbol:           "AAPL",
			TdaTransactionID: 0,
			Quantity:         decimal.NewFromFloat(2),
			CostBasis:        decimal.NewFromFloat(100),
			Date:             time.Now(),
		})
		require.NoError(t, err)

		_, _, err = tiService.ProcessSellOrder(ctx, tx, trade.ProcessSellOrderInput{
			Symbol:    "AAPL",
			Quantity:  decimal.NewFromFloat(1),
			CostBasis: decimal.NewFromFloat(110),
			Date:      time.Now(),
			Custodian: model.CustodianType_Tda,
		})
		require.NoError(t, err)

		_, _, err = tiService.ProcessSellOrder(ctx, tx, trade.ProcessSellOrderInput{
			Symbol:    "AAPL",
			Quantity:  decimal.NewFromFloat(1),
			CostBasis: decimal.NewFromFloat(130),
			Date:      time.Now(),
			Custodian: model.CustodianType_Tda,
		})
		require.NoError(t, err)

		out, err := CalculateNetRealizedReturns(tx)
		require.NoError(t, err)
		require.True(t, out.Equal(decimal.NewFromFloat(0.2)))
	})

	t.Run("stock split", func(t *testing.T) {
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade.ProcessTdaBuyOrderInput{
			Symbol:           "AAPL",
			TdaTransactionID: 0,
			Quantity:         decimal.NewFromFloat(1),
			CostBasis:        decimal.NewFromFloat(400),
			Date:             time.Now(),
			Description:      nil,
		})
		require.NoError(t, err)

		_, _, err = tiService.AddAssetSplit(ctx, tx, model.AssetSplit{
			Symbol:    "AAPL",
			Ratio:     4,
			Date:      time.Now(),
			CreatedAt: time.Now(),
		})
		require.NoError(t, err)

		_, _, err = tiService.ProcessSellOrder(ctx, tx, trade.ProcessSellOrderInput{
			Symbol:    "AAPL",
			Quantity:  decimal.NewFromFloat(4),
			CostBasis: decimal.NewFromFloat(110),
			Date:      time.Now(),
			Custodian: model.CustodianType_Tda,
		})
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

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade.ProcessTdaBuyOrderInput{
			Symbol:           "AAPL",
			TdaTransactionID: 0,
			Quantity:         decimal.NewFromFloat(1),
			CostBasis:        decimal.NewFromFloat(100),
			Date:             time.Now(),
			Description:      nil,
		})
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

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade.ProcessTdaBuyOrderInput{
			Symbol:           "AAPL",
			TdaTransactionID: 0,
			Quantity:         decimal.NewFromFloat(1),
			CostBasis:        decimal.NewFromFloat(100),
			Date:             time.Now(),
			Description:      nil,
		})
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

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade.ProcessTdaBuyOrderInput{
			Symbol:           "AAPL",
			TdaTransactionID: 0,
			Quantity:         decimal.NewFromFloat(1),
			CostBasis:        decimal.NewFromFloat(100),
			Date:             time.Now(),
			Description:      nil,
		})
		require.NoError(t, err)

		_, _, err = tiService.ProcessSellOrder(ctx, tx, trade.ProcessSellOrderInput{
			Symbol:    "AAPL",
			Quantity:  decimal.NewFromFloat(1),
			CostBasis: decimal.NewFromFloat(120),
			Date:      time.Now(),
			Custodian: model.CustodianType_Tda,
		})
		require.NoError(t, err)

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade.ProcessTdaBuyOrderInput{
			Symbol:           "AAPL",
			TdaTransactionID: 1,
			Quantity:         decimal.NewFromFloat(1),
			CostBasis:        decimal.NewFromFloat(100),
			Date:             time.Now(),
			Description:      nil,
		})
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
