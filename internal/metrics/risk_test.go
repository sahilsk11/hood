package metrics

import (
	"context"
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	. "hood/internal/domain"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func setupHistoricPrices(t *testing.T, tx *sql.Tx) {
	ctx := context.Background()
	start := time.Now()
	values := map[string][]float64{
		"SPY": {
			380, 382, 384, 380, 381, 385, 390, 392, 400,
			390, 389, 390, 396, 400, 402, 404, 400, 400, 410, 415,
		},
		"AAPL": {
			180.33, 182.01, 176.28, 178.96, 178.44, 166.02, 151.21,
			162.51, 174.55, 163.43, 155.74, 151.29, 148.31, 145.93,
			155.33, 164.9, 169.68, 177.3, 193.97, 192.46,
		},
	}
	prices := []model.Price{}
	for symbol, val := range values {
		for i, v := range val {
			prices = append(prices, model.Price{
				Symbol: symbol,
				Date:   start.AddDate(0, 0, -1*(len(val)-i-1)),
				Price:  decimal.NewFromFloat(v),
			})
		}
	}
	_, err := db.AddPrices(ctx, tx, prices)
	require.NoError(t, err)
}

func TestDailyStdevOfAsset(t *testing.T) {
	t.Run("SPY", func(t *testing.T) {
		tx := db.SetupTestDb(t)
		setupHistoricPrices(t, tx)

		stdev, err := DailyStdevOfAsset(tx, "SPY")
		require.NoError(t, err)

		require.InDelta(t, 0.01147232, stdev, 0.0001)
	})
	t.Run("AAPL", func(t *testing.T) {
		tx := db.SetupTestDb(t)
		setupHistoricPrices(t, tx)

		stdev, err := DailyStdevOfAsset(tx, "AAPL")
		require.NoError(t, err)
		require.InDelta(t, 0.053847922, stdev, 0.0001)
	})
}

func TestDailyStdevOfPortfolio(t *testing.T) {
	t.Run("SPY", func(t *testing.T) {
		tx := db.SetupTestDb(t)
		setupHistoricPrices(t, tx)

		portfolio := Portfolio{
			OpenLots: map[string][]*OpenLot{
				"SPY": {
					{
						Quantity: decimal.NewFromInt(10),
					},
				},
			},
			ClosedLots: map[string][]ClosedLot{},
		}
		stdev, err := DailyStdevOfPortfolio(tx, portfolio)
		require.NoError(t, err)

		require.InDelta(t, 0.01147232, stdev, 0.0001)
	})

	t.Run("multi-asset", func(t *testing.T) {
		tx := db.SetupTestDb(t)
		setupHistoricPrices(t, tx)

		portfolio := Portfolio{
			OpenLots: map[string][]*OpenLot{
				"SPY": {
					{
						Quantity: decimal.NewFromInt(1),
					},
				},
				"AAPL": {
					{
						Quantity: decimal.NewFromInt(2),
					},
				},
			},
			ClosedLots: map[string][]ClosedLot{},
		}
		stdev, err := DailyStdevOfPortfolio(tx, portfolio)
		require.NoError(t, err)

		require.InDelta(t, 0.0265872, stdev, 0.00001)
	})
}

func Test_setDifference(t *testing.T) {
	type args struct {
		s1 map[string]struct{}
		s2 map[string]struct{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "second set empty",
			args: args{
				s1: map[string]struct{}{"AAPL": {}, "DIS": {}},
				s2: map[string]struct{}{},
			},
			want: []string{"AAPL", "DIS"},
		},
		{
			name: "first set empty",
			args: args{
				s1: map[string]struct{}{},
				s2: map[string]struct{}{"AAPL": {}, "DIS": {}},
			},
			want: []string{"AAPL", "DIS"},
		},
		{
			name: "some overlap",
			args: args{
				s1: map[string]struct{}{"AAPL": {}},
				s2: map[string]struct{}{"AAPL": {}, "DIS": {}},
			},
			want: []string{"DIS"},
		},
		{
			name: "no difference",
			args: args{
				s1: map[string]struct{}{"AAPL": {}, "DIS": {}},
				s2: map[string]struct{}{"AAPL": {}, "DIS": {}},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setDifference(tt.args.s1, tt.args.s2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setDifference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func standardDeviation(data []float64) float64 {
	mean := mean(data)
	variance := 0.0

	for _, num := range data {
		variance += math.Pow(num-mean, 2)
	}

	variance = variance / float64(len(data))
	stdDev := math.Sqrt(variance)
	return stdDev
}

func mean(data []float64) float64 {
	sum := 0.0
	for _, num := range data {
		sum += num
	}
	return sum / float64(len(data))
}

func dstandardDeviation(data []decimal.Decimal) decimal.Decimal {
	mean := dmean(data)
	variance := decimal.Zero

	for _, num := range data {
		diff := num.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}

	variance = variance.Div(decimal.NewFromInt(int64(len(data))))
	stdDev := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))
	return stdDev
}

func dmean(data []decimal.Decimal) decimal.Decimal {
	sum := decimal.Zero
	for _, num := range data {
		sum = sum.Add(num)
	}
	return sum.Div(decimal.NewFromInt(int64(len(data))))
}
