package domain

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

type Portfolio struct {
	OpenLots   map[string][]*OpenLot
	ClosedLots map[string][]ClosedLot
	Cash       decimal.Decimal
	LastAction time.Time

	NewOpenLots []OpenLot // should be deprecated
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

// should be used sparingly - any operations on
// this portfolio are invalid
func (p1 Portfolio) Add(p2 Portfolio) Portfolio {
	p := p1.DeepCopy()
	p.Cash = p.Cash.Add(p2.Cash)

	for k, v := range p2.OpenLots {
		if _, ok := p.OpenLots[k]; !ok {
			p.OpenLots[k] = []*OpenLot{}
		}
		for _, x := range v {
			p.OpenLots[k] = append(p.OpenLots[k], x.DeepCopy())
		}
	}
	for k, v := range p.ClosedLots {
		if _, ok := p.ClosedLots[k]; !ok {
			p.ClosedLots[k] = []ClosedLot{}
		}
		for _, x := range v {
			p.ClosedLots[k] = append(p.ClosedLots[k], x.DeepCopy())
		}
	}
	for _, lot := range p.NewOpenLots {
		p.NewOpenLots = append(p.NewOpenLots, lot)
	}

	return p
}

type HistoricPortfolio struct {
	portfolios []Portfolio
}

func (hp HistoricPortfolio) GetPortfolios() []Portfolio {
	return hp.portfolios
}

func NewHistoricPortfolio() HistoricPortfolio {
	return HistoricPortfolio{
		portfolios: []Portfolio{},
	}
}

func (hp *HistoricPortfolio) sort() {
	sort.Slice(hp.portfolios, func(i, j int) bool {
		return hp.portfolios[i].LastAction.Before(hp.portfolios[j].LastAction)
	})
}

func (hp HistoricPortfolio) OnDate(t time.Time) Portfolio {
	i := 0
	latest := hp.portfolios[i]
	for i < len(hp.portfolios) && t.Before(hp.portfolios[i].LastAction) {
		i += 1
	}
	return latest
}

func (hp *HistoricPortfolio) Append(p Portfolio) {
	hp.portfolios = append(hp.portfolios, p)
	hp.sort()
}

func (hp HistoricPortfolio) Latest() Portfolio {
	return hp.portfolios[len(hp.portfolios)-1]
}
