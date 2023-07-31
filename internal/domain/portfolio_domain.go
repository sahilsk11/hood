package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Portfolio struct {
	OpenLots   map[string][]*OpenLot
	ClosedLots map[string][]ClosedLot
	Cash       decimal.Decimal
	LastAction time.Time

	NewOpenLots []OpenLot
}

func (p Portfolio) GetQuantity(symbol string) decimal.Decimal {
	lots := p.OpenLots[symbol]
	totalQuantity := decimal.Zero
	for _, lot := range lots {
		totalQuantity = totalQuantity.Add(lot.Quantity)
	}
	return totalQuantity
}

func (p Portfolio) DeepCopy() Portfolio {
	out := Portfolio{
		OpenLots:   map[string][]*OpenLot{},
		ClosedLots: map[string][]ClosedLot{},
		Cash:       p.Cash,
		LastAction: p.LastAction,
	}
	for k, v := range p.OpenLots {
		t := []*OpenLot{}
		for _, x := range v {
			t = append(t, x.DeepCopy())
		}
		out.OpenLots[k] = t
	}
	for k, v := range p.ClosedLots {
		t := []ClosedLot{}
		for _, x := range v {
			t = append(t, x.DeepCopy())
		}
		out.ClosedLots[k] = t
	}
	for _, lot := range p.NewOpenLots {
		out.NewOpenLots = append(out.NewOpenLots, lot)
	}

	return out
}

func (p Portfolio) GetOpenLotSymbols() []string {
	symbolsMap := map[string]struct{}{}
	for s := range p.OpenLots {
		symbolsMap[s] = struct{}{}
	}
	symbols := []string{}
	for s := range symbolsMap {
		symbols = append(symbols, s)
	}
	return symbols
}
