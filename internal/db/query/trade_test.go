package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	hood_errors "hood/internal"
	"hood/internal/db/models/postgres/public/model"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func Test_findDuplicateRhTrades(t *testing.T) {
	ctx := context.Background()
	connStr := "postgresql://postgres:postgres@localhost:5438/postgres_test?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	tx, err := db.Begin()
	require.NoError(t, err)

	tDate := time.Now()
	trades := []*model.Trade{
		{
			Symbol:      "AAPL",
			Action:      "BUY",
			Quantity:    decimal.NewFromFloat(10.5),
			CostBasis:   decimal.NewFromFloat(100.25),
			Date:        tDate,
			Description: nil,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Custodian:   model.CustodianType_Robinhood,
		},
	}

	_, err = AddTrades(ctx, tx, trades)
	require.NoError(t, err)

	err = findDuplicateRhTrades(tx, trades)
	fmt.Println(err)
	t.Fail()
	require.True(t, errors.As(err, &hood_errors.ErrDuplicateTrade{}), err)
}
