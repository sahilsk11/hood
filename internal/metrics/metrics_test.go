package metrics

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	. "hood/internal/domain"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

func Test_netValue(t *testing.T) {
	t.Run("only cash", func(t *testing.T) {
		p := Portfolio{
			OpenLots: map[string][]*domain.OpenLot{},
			Cash:     dec(100),
		}
		priceMap := map[string]decimal.Decimal{}
		result, err := netValue(p, priceMap)
		require.NoError(t, err)
		require.Equal(
			t,
			float64(100),
			result.InexactFloat64(),
		)
	})
	t.Run("few open lots", func(t *testing.T) {
		p := Portfolio{
			OpenLots: map[string][]*domain.OpenLot{
				"AAPL": {
					{
						Quantity: dec(100),
					},
				},
				"GOOG": {
					{
						Quantity: dec(1),
					},
				},
			},
			Cash: dec(100),
		}
		priceMap := map[string]decimal.Decimal{
			"AAPL": dec(1),
			"GOOG": dec(500),
		}
		result, err := netValue(p, priceMap)
		require.NoError(t, err)
		require.Equal(
			t,
			float64(700),
			result.InexactFloat64(),
		)
	})
}

func Test_TimeWeightedReturns(t *testing.T) {
	t.Run("simple", func(t *testing.T) {

		dailyPortfolioValues := map[string]decimal.Decimal{
			"2023-07-18": dec(100),
			"2023-07-19": dec(110),
		}

		out, err := DailyAggregateTwr(dailyPortfolioValues, map[string]decimal.Decimal{})
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

	t.Run("cash outflows", func(t *testing.T) {
		dailyPortfolio := map[string]decimal.Decimal{
			"2023-07-18": dec(100),
			"2023-07-19": dec(100),
		}
		transfers := map[string]decimal.Decimal{
			"2023-07-19": dec(-50),
		}

		out, err := DailyAggregateTwr(dailyPortfolio, transfers)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				map[string]decimal.Decimal{
					"2023-07-19": dec(1),
				},
				out,
			),
		)
	})

	t.Run("cash inflows", func(t *testing.T) {
		// TODO - investigate why this fails with
		// only two days

		dailyPortfolio := map[string]decimal.Decimal{
			"2023-07-18": dec(100),
			"2023-07-19": dec(110),
			"2023-07-20": dec(210),
		}
		transfers := map[string]decimal.Decimal{
			"2023-07-20": dec(100),
		}

		out, err := DailyAggregateTwr(dailyPortfolio, transfers)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				map[string]decimal.Decimal{
					"2023-07-19": dec(0.1),
					"2023-07-20": dec(0.1),
				},
				out,
			),
		)
	})

	t.Run("gains, losses, gains", func(t *testing.T) {
		dailyPortfolio := map[string]decimal.Decimal{
			"2023-07-18": dec(100),
			"2023-07-19": dec(150),
			"2023-07-20": dec(200),
			"2023-07-21": dec(400),
			"2023-07-22": dec(10),
			"2023-07-23": dec(150),
		}
		transfers := map[string]decimal.Decimal{}

		out, err := DailyAggregateTwr(dailyPortfolio, transfers)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				map[string]decimal.Decimal{
					"2023-07-19": dec(0.5),
					"2023-07-20": dec(1),
					"2023-07-21": dec(3),
					"2023-07-22": dec(-0.9),
					"2023-07-23": dec(0.5),
				},
				out,
				cmp.Comparer(func(x, y decimal.Decimal) bool {
					return (x.Sub(y)).Abs().LessThan(dec(0.0000000001))
				}),
			),
		)
	})

	t.Run("gains, losses, gains with cash flows", func(t *testing.T) {
		dailyPortfolio := map[string]decimal.Decimal{
			"2023-07-18": dec(100),
			"2023-07-19": dec(250),
			"2023-07-20": dec(250),
		}
		transfers := map[string]decimal.Decimal{
			"2023-07-19": dec(100),
			"2023-07-20": dec(-50),
		}

		out, err := DailyAggregateTwr(dailyPortfolio, transfers)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				map[string]decimal.Decimal{
					"2023-07-19": dec(0.25),
					"2023-07-20": dec(0.5625),
				},
				out,
				cmp.Comparer(func(x, y decimal.Decimal) bool {
					return (x.Sub(y)).Abs().LessThan(dec(0.0000000001))
				}),
			),
		)
	})

}

func TestDailyPortfolioValues(t *testing.T) {
	ctx := context.Background()
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	tx, err := dbConn.Begin()
	require.NoError(t, err)

	_, err = db.AddPrices(ctx, tx, []model.Price{
		{
			Symbol: "AAPL",
			Price:  dec(100),
			Date:   time.Date(2020, 02, 02, 0, 0, 0, 0, time.UTC),
		},
		{
			Symbol: "AAPL",
			Price:  dec(200),
			Date:   time.Date(2020, 02, 03, 0, 0, 0, 0, time.UTC),
		},
	})
	require.NoError(t, err)

	portfolios := map[string]Portfolio{
		"2020-02-02": {
			OpenLots: map[string][]*OpenLot{
				"AAPL": {
					{
						Quantity: dec(10),
					},
				},
			},
			Cash: dec(200),
		},
	}

	end := time.Date(2020, 02, 03, 0, 0, 0, 0, time.UTC)
	values, err := DailyPortfolioValues(
		tx,
		portfolios,
		nil,
		&end,
	)
	require.NoError(t, err)

	require.Equal(
		t,
		"",
		cmp.Diff(
			map[string]decimal.Decimal{
				"2020-02-02": dec(1200),
				"2020-02-03": dec(2200),
			},
			values,
		),
	)
}
