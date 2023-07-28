package metrics

import (
	db "hood/internal/db/query"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestStandardDeviation(t *testing.T) {
	t.Run("no trading holidays", func(t *testing.T) {
		dbConn, err := db.NewTest()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		dailyChange := map[string]decimal.Decimal{
			"1": dec(.2),
			"2": dec(.4),
			"3": dec(-0.1),
		}
		stdev, err := StandardDeviation(tx, dailyChange)
		require.NoError(t, err)

		require.InDelta(t, 0.205480466, stdev, 0.001)
	})
}
