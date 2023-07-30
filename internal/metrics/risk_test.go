package metrics

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestStandardDeviation(t *testing.T) {
	t.Run("no trading holidays", func(t *testing.T) {
		dbConn, err := db.NewTest()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		dailyChange := map[string]decimal.Decimal{
			"1": dec(.2),
			"2": dec(.4),
			"3": dec(-0.1),
		}
		stdev, err := StandardDeviation(tx, dailyChange)
		require.NoError(t, err)

		require.InDelta(t, 0.205480466, stdev, 0.001)
	})
	t.Run("trading holidays", func(t *testing.T) {
		ctx := context.Background()
		dbConn, err := db.NewTest()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		_, err = db.AddPrices(ctx, tx, []model.Price{
			{
				Symbol: "AAPL",
				Price:  dec(10),
				Date:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				Symbol: "AAPL",
				Price:  dec(10),
				Date:   time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
			},
		})
		require.NoError(t, err)
		dailyChange := map[string]decimal.Decimal{
			"2020-01-01": dec(.2),
			"2020-01-02": dec(1),
			"2020-01-03": dec(.2),
		}
		stdev, err := StandardDeviation(tx, dailyChange)
		require.NoError(t, err)

		require.Equal(t, float64(0), stdev)
	})
}
