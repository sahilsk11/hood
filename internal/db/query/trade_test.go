package db

import (
	"context"
	"errors"
	hood_errors "hood/internal"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func Test_findDuplicateRhTrades(t *testing.T) {
	ctx := context.Background()
	dbConn, err := NewTest()
	require.NoError(t, err)
	tx, err := dbConn.Begin()
	require.NoError(t, err)

	tDate := time.Now()
	trades := []domain.Trade{
		{
			Symbol:      "AAPL",
			Action:      "BUY",
			Quantity:    decimal.NewFromFloat(10.5),
			Price:       decimal.NewFromFloat(100.25),
			Date:        tDate,
			Description: nil,
			Custodian:   model.CustodianType_Robinhood,
		},
	}

	_, err = AddTrades(ctx, tx, trades)
	require.NoError(t, err)

	err = findDuplicateRhTrades(tx, tradesToDb(trades))
	require.True(t, errors.As(err, &hood_errors.ErrDuplicateTrade{}), err)
}
