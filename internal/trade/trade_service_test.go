package trade

import (
	"context"
	"errors"
	hood_errors "hood/internal"
	db "hood/internal/db/query"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func Test_tradeIngestionHandler_ProcessTdaBuyOrder(t *testing.T) {
	t.Run("try on test db", func(t *testing.T) {
		ctx := context.Background()

		dbConn, err := db.NewTest()
		require.NoError(t, err)

		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		tiService := NewTradeIngestionService()

		input := ProcessTdaBuyOrderInput{
			Symbol:           "AAPL",
			TdaTransactionID: int64(1),
			Quantity:         decimal.NewFromFloat(10.5),
			CostBasis:        decimal.NewFromFloat(100.25),
			Date:             time.Now(),
			Description:      nil,
		}
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, input)
		require.NoError(t, err)

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, input)

		require.True(t, errors.As(err, &hood_errors.ErrDuplicateTrade{}), err)
	})
}
