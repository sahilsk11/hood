package metrics

import (
	"hood/internal/domain"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

func TestPortfolio_netValue(t *testing.T) {
	t.Run("only cash", func(t *testing.T) {
		p := Portfolio{
			OpenLots: map[string][]*domain.OpenLot{},
			Transfer: dec(100),
		}
		priceMap := map[string]decimal.Decimal{}
		result, err := p.netValue(priceMap)
		require.NoError(t, err)
		require.Equal(
			t,
			float64(100),
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

		out, err := TimeWeightedReturns(dailyPortfolioValues, map[string]decimal.Decimal{})
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

	t.Run("realized gains", func(t *testing.T) {
		dailyPortfolio := map[string]decimal.Decimal{
			"2023-07-18": dec(100),
			"2023-07-19": dec(110),
		}

		out, err := TimeWeightedReturns(dailyPortfolio, map[string]decimal.Decimal{})
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

		out, err := TimeWeightedReturns(dailyPortfolio, transfers)
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

		out, err := TimeWeightedReturns(dailyPortfolio, transfers)
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

		out, err := TimeWeightedReturns(dailyPortfolio, transfers)
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
