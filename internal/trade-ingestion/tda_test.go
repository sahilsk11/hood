package trade_ingestion

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestParseTdaTransactionFile(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	tiService := NewMockTradeIngestionService(ctrl)

	tiService.
		EXPECT().
		ProcessTdaBuyOrder(ctx, model.Trade{
			Symbol:    "VTI",
			Action:    model.TradeActionType_Buy,
			Quantity:  decimal.NewFromFloat(2),
			CostBasis: decimal.NewFromFloat(191.12),
			Date:      time.Date(2023, 1, 6, 0, 0, 0, 0, time.UTC),
			Custodian: model.CustodianType_Tda,
		}, int64(47424103872))

	_, err := ParseTdaTransactionFile(ctx, "testdata/transactions.csv", tiService)
	require.NoError(t, err)
}
