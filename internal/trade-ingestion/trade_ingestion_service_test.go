package trade_ingestion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func Test_tradeIngestionHandler_ProcessTdaBuyOrder(t *testing.T) {
	t.Run("try on test db", func(t *testing.T) {
		ctx := context.Background()

		connStr := "postgresql://postgres:postgres@localhost:5438/postgres_test?sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		require.NoError(t, err)

		tx, err := db.Begin()
		require.NoError(t, err)
		tiService := NewTradeIngestionService(ctx, tx)

		tdaID := int64(1)
		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, model.Trade{
			Symbol:    "AAPL",
			Action:    "BUY",
			Quantity:  decimal.NewFromFloat(10.5),
			CostBasis: decimal.NewFromFloat(100.25),
			Date:      time.Now(),

			Description: nil,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Custodian:   model.CustodianType_Tda,
		}, tdaID)
		require.NoError(t, err, "failed on 2")
		fmt.Println("here")

		_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, model.Trade{
			Symbol:    "AAPL",
			Action:    "BUY",
			Quantity:  decimal.NewFromFloat(10.5),
			CostBasis: decimal.NewFromFloat(100.25),
			Date:      time.Now(),

			Description: nil,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Custodian:   model.CustodianType_Tda,
		}, tdaID)
		fmt.Println(err)

		require.True(t, errors.As(err, &ErrDuplicateTrade{}), err)
		tx.Rollback()
	})
}
