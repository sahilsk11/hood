package domain

import (
	"github.com/shopspring/decimal"
)

type Portfolio struct {
	OpenLots   map[string][]*OpenLot
	ClosedLots map[string][]ClosedLot
	Cash       decimal.Decimal
}

func (pt Portfolio) ToHoldings() *Holdings {
	out := &Holdings{
		Positions: map[string]*Position{},
		Cash:      pt.Cash,
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

type Position struct {
	Symbol         string
	Quantity       decimal.Decimal
	TotalCostBasis decimal.Decimal
}

type Holdings struct {
	Positions map[string]*Position
	Cash      decimal.Decimal
}

func (mp Holdings) Symbols() []string {
	out := []string{}
	for symbol := range mp.Positions {
		out = append(out, symbol)
	}
	return out
}
