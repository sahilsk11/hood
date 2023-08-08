package metrics

import (
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

func TestDailyPortfolioValues(t *testing.T) {
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	tx, err := dbConn.Begin()
	require.NoError(t, err)

	_, err = db.AddPrices(tx, []model.Price{
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

	portfolios := HistoricPortfolio{}
	portfolios.Append(Portfolio{
		OpenLots: map[string][]*OpenLot{
			"AAPL": {
				{
					Quantity: dec(10),
				},
			},
		},
		Cash: dec(200),
	})

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
