package trade

import (
	"context"
	db "hood/internal/db/query"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestParseTdaTransactionFile(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	tx, err := dbConn.Begin()
	require.NoError(t, err)
	db.RollbackAfterTest(t, tx)

	tiService := NewMockTradeIngestionService(ctrl)

	tiService.
		EXPECT().
		ProcessTdaBuyOrder(ctx, tx, ProcessTdaBuyOrderInput{
			Symbol:           "VTI",
			Quantity:         decimal.NewFromFloat(2),
			CostBasis:        decimal.NewFromFloat(191.12),
			Date:             time.Date(2023, 1, 6, 0, 0, 0, 0, time.UTC),
			TdaTransactionID: int64(47424103872),
		})

	_, err = ParseTdaTransactionFile(ctx, tx, "testdata/transactions.csv", tiService)
	require.NoError(t, err)
}
