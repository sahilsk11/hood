package metrics

import (
	"context"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/trade"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func TestDailyReturns(t *testing.T) {
	endTime := time.Now()
	t.Run("single buy and sell", func(t *testing.T) {
		startTime := time.Date(2020, 06, 19, 0, 0, 0, 0, time.UTC)
		trades := []model.Trade{
			{
				Symbol:    "AAPL",
				Quantity:  dec(2),
				CostBasis: dec(100),
				Date:      time.Date(2020, 06, 19, 0, 0, 0, 0, time.UTC),
				Action:    model.TradeActionType_Buy,
			},
			{
				Symbol:    "AAPL",
				Quantity:  dec(1),
				CostBasis: dec(110),
				Date:      time.Date(2020, 06, 19, 0, 0, 0, 0, time.UTC),
				Action:    model.TradeActionType_Sell,
			},
		}
		assetSplits := []model.AssetSplit{}
		out, err := DailyReturns(trades, assetSplits, startTime, endTime)

		require.Equal(t,
			"",
			cmp.Diff(
				map[string]Portfolio{
					"2020-06-19": {
						OpenLots: map[string][]*domain.OpenLot{
							"AAPL": {
								{
									OpenLotID: 0,
									Quantity:  dec(1),
									CostBasis: dec(100),
									Trade:     trades[0],
								},
							},
						},
						ClosedLots: map[string][]*domain.ClosedLot{
							"AAPL": {
								{
									SellTrade:     trades[1],
									Quantity:      dec(1),
									RealizedGains: dec(10),
								},
							},
						},
					},
				},
				out,
				cmpopts.IgnoreFields(domain.ClosedLot{}, "GainsType"),
			),
		)

		require.NoError(t, err)
	})

	t.Run("close open lot", func(t *testing.T) {
		startTime := time.Date(2020, 06, 19, 0, 0, 0, 0, time.UTC)
		trades := []model.Trade{
			{
				Symbol:    "AAPL",
				Quantity:  dec(1),
				CostBasis: dec(100),
				Date:      time.Date(2020, 06, 19, 0, 0, 0, 0, time.UTC),
				Action:    model.TradeActionType_Buy,
			},
			{
				Symbol:    "AAPL",
				Quantity:  dec(1),
				CostBasis: dec(110),
				Date:      time.Date(2020, 06, 19, 0, 0, 0, 0, time.UTC),
				Action:    model.TradeActionType_Sell,
			},
		}
		assetSplits := []model.AssetSplit{}
		out, err := DailyReturns(trades, assetSplits, startTime, endTime)

		require.Equal(t,
			"",
			cmp.Diff(
				map[string]Portfolio{
					"2020-06-19": {
						OpenLots: map[string][]*domain.OpenLot{},
						ClosedLots: map[string][]*domain.ClosedLot{
							"AAPL": {
								{
									SellTrade:     trades[1],
									Quantity:      dec(1),
									RealizedGains: dec(10),
								},
							},
						},
					},
				},
				out,
				cmpopts.IgnoreFields(domain.ClosedLot{}, "GainsType"),
			),
		)

		require.NoError(t, err)
	})

	t.Run("real thing", func(t *testing.T) {
		// t.Skip()
		startTime := time.Date(2020, 06, 19, 0, 0, 0, 0, time.UTC)

		dbConn, err := db.New()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)

		trades, err := db.GetHistoricTrades(tx)
		require.NoError(t, err)
		assetSplits, err := db.GetHistoricAssetSplits(tx)
		require.NoError(t, err)

		out, err := DailyReturns(trades, assetSplits, startTime, endTime)
		require.NoError(t, err)

		dStr := "2023-07-10"
		d, err := time.Parse("2006-01-02", dStr)
		require.NoError(t, err)
		prices, err := db.GetPricesOnDate(tx, d, []string{})
		require.NoError(t, err)
		p, ok := out[dStr]
		if !ok {
			t.Fatalf("missing date")
		}

		// bytes, err := json.Marshal(p)
		// require.NoError(t, err)
		// fmt.Println(string(bytes))

		fmt.Println(p.CalculateReturns(prices))
		t.Fail()
	})
}

func dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

func TestPortfolio_CalculateReturns(t *testing.T) {
	t.Run("only closed lots", func(t *testing.T) {
		p := Portfolio{
			OpenLots: map[string][]*domain.OpenLot{},
			ClosedLots: map[string][]*domain.ClosedLot{
				"AAPL": {
					{
						Quantity:      dec(1),
						RealizedGains: dec(10),
						SellTrade: model.Trade{
							CostBasis: dec(110),
						},
					},
				},
				"GOOG": {
					{
						Quantity:      dec(1),
						RealizedGains: dec(10),
						SellTrade: model.Trade{
							CostBasis: dec(110),
						},
					},
				},
				"META": {
					{
						Quantity:      dec(1),
						RealizedGains: dec(-5),
						SellTrade: model.Trade{
							CostBasis: dec(95),
						},
					},
				},
			},
		}
		priceMap := map[string]decimal.Decimal{}
		result, err := p.CalculateReturns(priceMap)
		require.NoError(t, err)
		require.Equal(
			t,
			0.05,
			result.InexactFloat64(),
		)
	})
}
