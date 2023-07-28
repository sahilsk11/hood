package metrics

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func Test_DailyAggregateTwr(t *testing.T) {
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
					"2023-07-19": dec(1.1),
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
					"2023-07-19": dec(2),
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
					"2023-07-19": dec(1.1),
					"2023-07-20": dec(1.1),
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
					"2023-07-19": dec(1.5),
					"2023-07-20": dec(2),
					"2023-07-21": dec(4),
					"2023-07-22": dec(0.1),
					"2023-07-23": dec(1.5),
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
					"2023-07-19": dec(1.25),
					"2023-07-20": dec(1.5625),
				},
				out,
				cmp.Comparer(func(x, y decimal.Decimal) bool {
					return (x.Sub(y)).Abs().LessThan(dec(0.0000000001))
				}),
			),
		)
	})

}
