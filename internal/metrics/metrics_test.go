package metrics

import (
	"context"
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
			Quantity:         dec(1),
			CostBasis:        dec(100),
			Date:             time.Now(),
			Description:      nil,
		})
		require.NoError(t, err)

		_, err = db.AddPrices(ctx, tx, []model.Price{
			{
				Symbol:    "AAPL",
				Price:     dec(100),
				UpdatedAt: time.Now(),
			},
		})
		require.NoError(t, err)

		out, err := CalculateNetUnrealizedReturns(tx)
		require.NoError(t, err)
		require.True(t, out.Equal(decimal.Zero), out)
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

func TestDailyPortfolio(t *testing.T) {
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
		out, err := DailyPortfolio(trades, assetSplits, nil, startTime, endTime)

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
									Trade:     &trades[0],
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
		out, err := DailyPortfolio(trades, assetSplits, nil, startTime, endTime)

		require.Equal(t,
			"",
			cmp.Diff(
				map[string]Portfolio{
					"2020-06-19": {
						OpenLots: map[string][]*domain.OpenLot{},
					},
				},
				out,
				cmpopts.IgnoreFields(domain.ClosedLot{}, "GainsType"),
			),
		)

		require.NoError(t, err)
	})
}

func dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

func TestPortfolio_CalculatePortfolioValue(t *testing.T) {
	t.Run("only closed lots", func(t *testing.T) {
		p := Portfolio{
			OpenLots: map[string][]*domain.OpenLot{},
		}
		priceMap := map[string]decimal.Decimal{}
		result, err := p.CalculatePortfolioValue(priceMap)
		require.NoError(t, err)
		require.Equal(
			t,
			0.05,
			result.InexactFloat64(),
		)
	})
}

func TestDailyReturns(t *testing.T) {
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	t.Run("simple", func(t *testing.T) {
		ctx := context.Background()
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		d, err := time.Parse("2006-01-02", "2023-07-19")
		require.NoError(t, err)

		_, err = db.AddPrices(ctx, tx, []model.Price{
			{
				Symbol: "AAPL",
				Price:  dec(100),
				Date:   d.AddDate(0, 0, -1),
			},
			{
				Symbol: "AAPL",
				Price:  dec(110),
				Date:   d,
			},
		})
		require.NoError(t, err)

		transfer := &model.BankActivity{
			Date:         d.AddDate(0, 0, -1),
			Amount:       dec(100),
			ActivityType: model.BankActivityType_Deposit,
		}
		err = db.AddTransfer(tx, transfer)
		require.NoError(t, err)

		dailyPortfolio := map[string]Portfolio{
			"2023-07-18": {
				OpenLots: map[string][]*domain.OpenLot{
					"AAPL": {
						{
							Quantity:     dec(1),
							CostBasis:    dec(100),
							PurchaseDate: d.AddDate(0, 0, -1),
						},
					},
				},
				// Cash: dec(100),
			},
			"2023-07-19": {
				OpenLots: map[string][]*domain.OpenLot{
					"AAPL": {
						{
							Quantity:     dec(1),
							CostBasis:    dec(100),
							PurchaseDate: d,
						},
					},
				},
				// Cash: dec(100),
			},
		}

		out, err := DailyReturns(tx, dailyPortfolio)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				map[string]decimal.Decimal{
					"2023-07-19": dec(0.1),
				},
				out,
			),
		)
	})

	t.Run("some gains, some net even", func(t *testing.T) {
		ctx := context.Background()
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		d, err := time.Parse("2006-01-02", "2023-07-14")
		require.NoError(t, err)
		db.AddPrices(ctx, tx, []model.Price{
			{
				Symbol: "AAPL",
				Price:  dec(110),
				Date:   d,
			},
		})

		dailyPortfolio := map[string]Portfolio{
			"2023-07-15": {
				OpenLots: map[string][]*domain.OpenLot{
					"AAPL": {
						{
							Quantity:  dec(1),
							CostBasis: dec(110),
						},
					},
				},
			},
		}

		out, err := DailyReturns(tx, dailyPortfolio)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				map[string]decimal.Decimal{
					"2023-07-15": dec(0.1),
				},
				out,
			),
			out["2023-01-01"].String())
	})

}
