package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Portfolio is the best representation of a user's
// trading account. has essential information
type Portfolio struct {
	OpenLots   map[string][]*OpenLot
	ClosedLots map[string][]ClosedLot
	Cash       decimal.Decimal // deprecating until i figure it out
}

// Holdings is a simplified version of a the portfolio. We
// ignore cost basis and closed lots, and only look at
// what they hold today in aggregate. FKA MetricsPortfolio
// IMO "AggregatePortfolio" was not a great name
type Holdings struct {
	Positions map[string]*Position
	Cash      decimal.Decimal
}

// Position represents a set of open lots under
// a single symbol. It's used as the underlying
// domain for Holdings, so it only contains aggregate
// data
type Position struct {
	Symbol         string
	Quantity       decimal.Decimal
	TotalCostBasis decimal.Decimal
}

// HistoricPortfolio represents the history of a portfolio
// with full detail
type HistoricPortfolio []PortfolioOnDate
type PortfolioOnDate struct {
	Portfolio Portfolio
	Date      time.Time
}

func (hp HistoricPortfolio) Latest() *Portfolio {
	if len(hp) == 0 {
		return nil
	}
	return &hp[len(hp)-1].Portfolio
}

func (pt Portfolio) ToHoldings() *Holdings {
	out := &Holdings{
		Positions: map[string]*Position{},
		// Cash:      pt.Cash,
	}
	for symbol, lots := range pt.OpenLots {
		totalQuantity := decimal.Zero
		for _, lot := range lots {
			totalQuantity = totalQuantity.Add(lot.Quantity)
		}
		out.Positions[symbol] = &Position{
			Symbol:   symbol,
			Quantity: totalQuantity,
		}
	}
	return out
}

func (mp Holdings) Symbols() []string {
	out := []string{}
	for symbol := range mp.Positions {
		out = append(out, symbol)
	}
	return out
}

func (p Portfolio) DeepCopy()
