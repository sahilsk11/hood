package trade_ingestion

import (
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProcessTrades(t *testing.T) {
	// t.Run("basic lot aggregation creates correct open lots", func(t *testing.T) {
	// 	tests := []struct {
	// 		name             string
	// 		outfileEntries           []entry
	// 		expectedOpenLots []*model.OpenLot
	// 	}{
	// 		{
	// 			"buy 2, sell 1",
	// 			[]entry{
	// 				{
	// 					TradeID:  439,
	// 					Action:   model.TradeActionType_Buy,
	// 					Quantity: decimal.NewFromInt(1),
	// 					Date:     newDate("2020-01-01"),
	// 				},
	// 				{
	// 					TradeID:  440,
	// 					Action:   model.TradeActionType_Buy,
	// 					Quantity: decimal.NewFromInt(1),
	// 					Date:     newDate("2021-01-01"),
	// 				},
	// 				{
	// 					TradeID:  441,
	// 					Action:   model.TradeActionType_Sell,
	// 					Quantity: decimal.NewFromInt(1),
	// 					Date:     newDate("2022-01-01"),
	// 				},
	// 			},
	// 			[]*model.OpenLot{
	// 				{
	// 					TradeID:  440,
	// 					Quantity: decimal.NewFromInt(1),
	// 				},
	// 			},
	// 		},
	// 		{
	// 			"partial deletes of lot",
	// 			[]model.Trade{
	// 				{
	// 					TradeID:  439,
	// 					Action:   model.TradeActionType_Buy,
	// 					Quantity: decimal.NewFromInt(5),
	// 					Date:     newDate("2020-01-01"),
	// 				},
	// 				{
	// 					TradeID:  440,
	// 					Action:   model.TradeActionType_Buy,
	// 					Quantity: decimal.NewFromInt(5),
	// 					Date:     newDate("2021-01-01"),
	// 				},
	// 				{
	// 					TradeID:  441,
	// 					Action:   model.TradeActionType_Sell,
	// 					Quantity: decimal.NewFromInt(8),
	// 					Date:     newDate("2022-01-01"),
	// 				},
	// 			},
	// 			[]*model.OpenLot{
	// 				{
	// 					TradeID:  440,
	// 					Quantity: decimal.NewFromInt(2),
	// 				},
	// 			},
	// 		},
	// 	}

	// 	for _, test := range tests {
	// 		out, err := ProcessHistoricTrades(test.trades, nil)
	// 		require.NoError(t, err)
	// 		assertOpenLotsEqual(t, test.expectedOpenLots, out.OpenLots)
	// 	}
	// })

}

func assertOpenLotsEqual(t *testing.T, expected, actual []*model.OpenLot) {
	actualWithoutDeleted := []*model.OpenLot{}
	for _, lot := range actual {
		if lot.DeletedAt == nil {
			actualWithoutDeleted = append(actualWithoutDeleted, lot)
		}
	}
	actual = actualWithoutDeleted

	if len(expected) != len(actual) {
		require.Fail(t, fmt.Sprintf("incorrect lot length: expected %d, got %d", len(expected), len(actual)))
	}
	for i := 0; i < len(expected); i++ {
		exp := expected[i]
		act := actual[i]
		msg := fmt.Sprintf("mismatched %dth element's ", i)
		require.Equal(t, exp.TradeID, act.TradeID, msg+"TradeID")
		require.Equal(t, exp.TradeID, act.TradeID, msg+"TradeID")
		require.Equal(t, exp.Quantity, act.Quantity, msg+"Quantity")
	}
}

// helper to write dates in one line
func newDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}
	return t
}

func Test_entryIterator(t *testing.T) {
	t.Run("ensure hasNext properly handles nil lists", func(t *testing.T) {
		i := newEntryiterator(nil, nil)
		require.False(t, i.hasNext())
	})

	t.Run("ensure hasNext properly handles empty lists", func(t *testing.T) {
		i := newEntryiterator([]*model.Trade{}, []*model.AssetSplit{})
		require.False(t, i.hasNext())
	})

	t.Run("ensure hasNext and next properly handles one trade and no splits", func(t *testing.T) {
		i := newEntryiterator([]*model.Trade{{}}, []*model.AssetSplit{})
		require.True(t, i.hasNext())
		nextTrade, nextSplit := i.next()
		require.NotNil(t, nextTrade)
		require.Nil(t, nextSplit)
	})

	t.Run("ensure hasNext and next properly handles one split and no trades", func(t *testing.T) {
		i := newEntryiterator([]*model.Trade{}, []*model.AssetSplit{{}})
		require.True(t, i.hasNext())
		nextTrade, nextSplit := i.next()
		require.NotNil(t, nextSplit)
		require.Nil(t, nextTrade)
	})

	t.Run("ensure hasNext and next iterates", func(t *testing.T) {
		i := newEntryiterator([]*model.Trade{{Date: time.Now()}}, []*model.AssetSplit{{Date: time.Now().Add(time.Hour)}})
		require.True(t, i.hasNext())
		nextTrade, nextSplit := i.next()
		require.NotNil(t, nextTrade)
		require.Nil(t, nextSplit)

		require.True(t, i.hasNext())
		nextTrade, nextSplit = i.next()
		require.NotNil(t, nextSplit)
		require.Nil(t, nextTrade)

		require.False(t, i.hasNext())
	})
}
