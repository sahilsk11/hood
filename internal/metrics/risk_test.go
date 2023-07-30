package metrics

import (
	"fmt"
	db "hood/internal/db/query"
	. "hood/internal/domain"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestDailyStdevOfPortfolio(t *testing.T) {
	t.Run("SPY", func(t *testing.T) {
		dbConn, err := db.New()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

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

		fmt.Println(stdev)
		t.Fail()
	})

	t.Run("multi-asset", func(t *testing.T) {
		dbConn, err := db.New()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		portfolio := Portfolio{
			OpenLots: map[string][]*OpenLot{
				"SPY": {
					{
						Quantity: decimal.NewFromInt(10),
					},
				},
				"AAPL": {
					{
						Quantity: decimal.NewFromInt(10),
					},
				},
			},
			ClosedLots: map[string][]ClosedLot{},
		}
		stdev, err := DailyStdevOfPortfolio(tx, portfolio)
		require.NoError(t, err)

		fmt.Println(stdev)
		t.Fail()
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

func TestDailyStdevOfAsset(t *testing.T) {
	t.Run("SPY", func(t *testing.T) {
		dbConn, err := db.New()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		stdev, err := DailyStdevOfAsset(tx, "SPY")
		require.NoError(t, err)

		fmt.Println(stdev)
		t.Fail()
	})
	t.Run("AAPL", func(t *testing.T) {
		dbConn, err := db.New()
		require.NoError(t, err)
		tx, err := dbConn.Begin()
		require.NoError(t, err)
		db.RollbackAfterTest(t, tx)

		stdev, err := DailyStdevOfAsset(tx, "AAPL")
		require.NoError(t, err)

		fmt.Println(stdev)
		t.Fail()
	})
}
