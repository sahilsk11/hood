package trade

import (
	"context"
	"errors"
	hood_errors "hood/internal"
	"hood/internal/db/models/postgres/public/model"
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

		dbConn, err := db.New()
		require.NoError(t, err)

		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		tiService := NewTradeIngestionService()

		tdaID := int64(1)
		trade := model.Trade{
			Symbol:      "AAPL",
			Action:      "BUY",
			Quantity:    decimal.NewFromFloat(10.5),
			CostBasis:   decimal.NewFromFloat(100.25),
			Date:        time.Now(),
			Description: nil,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Custodian:   model.CustodianType_Tda,
		}
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade, tdaID)
		require.NoError(t, err)

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, trade, tdaID)

		require.True(t, errors.As(err, &hood_errors.ErrDuplicateTrade{}), err)
	})
}
