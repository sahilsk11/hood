package portfolio

import (
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	. "hood/internal/domain"
	"hood/internal/util"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPlaybackDaily(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		times := []time.Time{
			time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 2, 2, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 3, 3, 0, 0, 0, time.UTC),
		}
		events := Events{
			Transfers: []domain.Transfer{{
				Amount: dec(100),
				Date:   times[0],
			}},
			Trades: []domain.Trade{
				{
					Action:   model.TradeActionType_Buy,
					Symbol:   "AAPL",
					Date:     times[1],
					Quantity: dec(1),
					Price:    dec(50),
				},
				{
					Action:   model.TradeActionType_Sell,
					Symbol:   "AAPL",
					Quantity: dec(1),
					Price:    dec(50),
					Date:     times[2],
				},
			},
		}
		dailyPortfolios, err := PlaybackDaily(events)
		require.NoError(t, err)
		require.Equal(
			t,
			"",
			cmp.Diff(
				map[string]Portfolio{
					"2020-01-01": {
						OpenLots:   map[string][]*OpenLot{},
						ClosedLots: map[string][]ClosedLot{},
						Cash:       dec(100),
						LastAction: times[0],
					},
					"2020-01-02": {
						OpenLots: map[string][]*OpenLot{
							"AAPL": {
								{
									Quantity:  dec(1),
									CostBasis: dec(50),
									Trade:     &events.Trades[0],
									Date:      times[1],
								},
							},
						},
						ClosedLots: map[string][]ClosedLot{},
						Cash:       dec(50),
						LastAction: times[1],
					},
					"2020-01-03": {
						OpenLots: map[string][]*OpenLot{},
						ClosedLots: map[string][]ClosedLot{
							"AAPL": {
								{
									RealizedGains: dec(0),
									GainsType:     model.GainsType_ShortTerm,
									Quantity:      dec(1),
									OpenLot: &domain.OpenLot{
										Quantity:  dec(0),
										CostBasis: dec(50),
										Trade:     &events.Trades[0],
										Date:      times[2],
									},
									SellTrade: &events.Trades[1],
								},
							},
						},
						Cash:       dec(100),
						LastAction: times[2],
						NewOpenLots: []domain.OpenLot{
							{
								Quantity:  dec(0),
								CostBasis: dec(50),
								Trade:     &events.Trades[0],
								Date:      times[2],
							},
						},
					},
				},
				dailyPortfolios,
			),
		)
	})

	t.Run("two sells on same day", func(t *testing.T) {
		times := []time.Time{
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		}
		events := Events{
			Trades: []Trade{
				{
					Symbol:   "AAPL",
					Action:   "BUY",
					Quantity: dec(2),
					Date:     times[0],
				},
				{
					Symbol:   "AAPL",
					Action:   "SELL",
					Quantity: dec(1),
					Date:     times[1],
				},
				{
					Symbol:   "AAPL",
					Action:   "SELL",
					Quantity: dec(1),
					Date:     times[1],
				},
			},
		}
		out, err := PlaybackDaily(events)
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				map[string]Portfolio{
					"2020-01-01": {
						OpenLots: map[string][]*OpenLot{
							"AAPL": {
								{},
							},
						},
						ClosedLots: map[string][]ClosedLot{},
						LastAction: times[0],
					},
					"2020-01-02": {
						OpenLots: map[string][]*OpenLot{},
						ClosedLots: map[string][]ClosedLot{
							"AAPL": {
								{},
								{},
							},
						},
						LastAction: times[1],
						NewOpenLots: []OpenLot{
							{},
							{},
						},
					},
				},
				out,
				cmp.Comparer(func(x, y *OpenLot) bool {
					return x == nil && y == nil || (x != nil && y != nil)
				}),
				cmp.Comparer(func(x, y []ClosedLot) bool {
					return len(x) == len(y)
				}),
				cmp.Comparer(func(x, y []OpenLot) bool {
					return len(x) == len(y)
				}),
			),
		)
	})

}

func TestPlayback(t *testing.T) {
	start := time.Now()
	times := []time.Time{
		start.Add(-3 * time.Second),
		start.Add(-2 * time.Second),
		start.Add(-1 * time.Second),
	}
	trades := []Trade{
		{
			Symbol:   "AAPL",
			Quantity: dec(10),
			Price:    dec(100),
			Action:   model.TradeActionType_Buy,
			Date:     times[1],
		},
		{
			Symbol:   "AAPL",
			Quantity: dec(10),
			Price:    dec(100),
			Action:   model.TradeActionType_Sell,
			Date:     times[2],
		},
	}
	transfers := []Transfer{{Amount: dec(1000), Date: times[0]}}
	events := Events{
		Trades:    trades,
		Transfers: transfers,
	}
	out, err := Playback(events)
	require.NoError(t, err)
	require.Equal(
		t,
		"",
		cmp.Diff(
			Portfolio{
				OpenLots: map[string][]*OpenLot{},
				ClosedLots: map[string][]ClosedLot{
					"AAPL": {
						{
							OpenLot: &OpenLot{
								Trade:     &trades[0],
								Quantity:  dec(0),
								CostBasis: dec(100),
								Date:      times[2],
							},
							Quantity:      dec(10),
							GainsType:     model.GainsType_ShortTerm,
							SellTrade:     &trades[1],
							RealizedGains: dec(0),
						},
					},
				},
				LastAction: times[2],
				Cash:       dec(1000),
				NewOpenLots: []domain.OpenLot{
					{
						Trade:     &trades[0],
						Quantity:  dec(0),
						CostBasis: dec(100),
						Date:      times[2],
					},
				},
			},
			*out,
		),
	)
}

func Test_insertPortfolio(t *testing.T) {
	dbConn, err := db.NewTest()
	require.NoError(t, err)
	tx, err := dbConn.Begin()
	require.NoError(t, err)

	times := []time.Time{
		time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
		time.Date(2020, 1, 2, 2, 0, 0, 0, time.UTC),
		time.Date(2020, 1, 3, 3, 0, 0, 0, time.UTC),
	}

	trades := []domain.Trade{
		{
			Symbol:    "AAPL",
			Quantity:  decimal.NewFromFloat(50),
			Price:     decimal.NewFromFloat(100),
			Date:      times[0],
			Custodian: model.CustodianType_Robinhood,
			Action:    model.TradeActionType_Buy,
		},
		{
			Symbol:    "AAPL",
			Quantity:  decimal.NewFromFloat(50),
			Price:     decimal.NewFromFloat(100),
			Date:      times[1],
			Custodian: model.CustodianType_Robinhood,
			Action:    model.TradeActionType_Sell,
		},
	}
	_, err = db.AddTrades(nil, tx, trades)
	require.NoError(t, err)

	tr, err := db.GetHistoricTrades(tx, model.CustodianType_Robinhood)
	require.NoError(t, err)

	events := Events{
		Trades: tr,
	}
	dailyPortfolios, err := PlaybackDaily(events)
	require.NoError(t, err)
	util.Pprint(dailyPortfolios)

	for _, portfolio := range dailyPortfolios {
		err = insertPortfolio(tx, portfolio)
		if err != nil {
			require.NoError(t, err)
		}
	}

	// tx.Commit()
}

func dec(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}
