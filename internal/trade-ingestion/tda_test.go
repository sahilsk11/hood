package trade_ingestion

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestParseTdaTransactionFile(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	tiService := NewMockTradeIngestionService(ctrl)

	_, err := ParseTdaTransactionFile(ctx, "testdata/transactions.csv", tiService)
	require.NoError(t, err)
}
