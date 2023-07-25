package trade

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
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
		ProcessTdaBuyOrder(ctx, tx, domain.Trade{
			Symbol:   "VTI",
			Quantity: decimal.NewFromFloat(2),
			Price:    decimal.NewFromFloat(191.12),
			Date:     time.Date(2023, 1, 6, 0, 0, 0, 0, time.UTC),
			Action:   model.TradeActionType_Buy,
		}, int64(47424103872))

	_, err = ParseTdaTransactionFile(ctx, tx, "testdata/transactions.csv", tiService)
	require.NoError(t, err)
}
