package portfolio

import (
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/domain"
	"hood/internal/trade"
	"hood/internal/util"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// replay historic events

func mergeEvents(in Events) []TradeEvent {
	out := []TradeEvent{}
	for _, t := range in.Trades {
		out = append(out, t)
	}
	for _, t := range in.AssetSplits {
		out = append(out, t)
	}
	for _, t := range in.Transfers {
		out = append(out, t)
	}
	for _, t := range in.Dividends {
		out = append(out, t)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].GetDate().Before(out[j].GetDate())
	})
	return out
}

type Events struct {
	Trades      []Trade
	AssetSplits []AssetSplit
	Transfers   []Transfer
	Dividends   []Dividend
}

func Playback(in Events) (*Portfolio, error) {
	daily, err := PlaybackDaily(in)
	if err != nil {
		return nil, err
	}
	keys := []string{}
	for k := range daily {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	latest := daily[keys[len(keys)-1]]

	return &latest, nil
}

// given historical events, calculate all portfolios
func PlaybackDaily(in Events) (map[string]Portfolio, error) {
	portfolio := &Portfolio{
		OpenLots:   map[string][]*OpenLot{},
		ClosedLots: map[string][]ClosedLot{},
	}
	mappedPortfolio := map[string]Portfolio{}

	events := mergeEvents(in)
	if len(events) == 0 {
		return nil, fmt.Errorf("no events found")
	}
	util.Pprint(events)

	for _, e := range events {
		switch e.(type) {
		case Trade:
			t := e.(Trade)
			if t.Action == model.TradeActionType_Buy {
				handleBuy(t, portfolio)
			} else if t.Action == model.TradeActionType_Sell {
				err := handleSell(t, portfolio)
				if err != nil {
					return nil, fmt.Errorf("failed to execute sell %v: %w", t, err)
				}
			}
		case AssetSplit:
			handleAssetSplit(e.(AssetSplit), portfolio)
		case Transfer:
			portfolio.Cash = portfolio.Cash.Add(e.(Transfer).Amount)
		case Dividend:
			portfolio.Cash = portfolio.Cash.Add(e.(Dividend).Amount)
		}
		date := e.GetDate().Format("2006-01-02")
		portfolio.LastAction = e.GetDate()
		mappedPortfolio[date] = portfolio.DeepCopy()
	}
	return mappedPortfolio, nil
}

func handleBuy(t Trade, p *Portfolio) {
	if _, ok := p.OpenLots[t.Symbol]; !ok {
		p.OpenLots[t.Symbol] = []*OpenLot{}
	}
	newLot := OpenLot{
		Trade:     &t,
		Quantity:  t.Quantity,
		CostBasis: t.Price,
		Date:      t.Date,
	}
	p.OpenLots[t.Symbol] = append(p.OpenLots[t.Symbol], &newLot)
	p.Cash = p.Cash.Sub(t.Price.Mul(t.Quantity))
}

func handleSell(t Trade, p *Portfolio) error {
	openLots := []*OpenLot{}
	if lots, ok := p.OpenLots[t.Symbol]; ok {
		openLots = lots
	}
	closedLots := []ClosedLot{}
	if lots, ok := p.ClosedLots[t.Symbol]; ok {
		closedLots = lots
	}
	result, err := trade.PreviewSellOrder(t, openLots)
	if err != nil {
		return err
	}

	p.Cash = p.Cash.Add(result.CashDelta)
	p.OpenLots[t.Symbol] = result.OpenLots
	if len(p.OpenLots[t.Symbol]) == 0 {
		delete(p.OpenLots, t.Symbol)
	}
	p.NewOpenLots = result.NewOpenLots

	p.ClosedLots[t.Symbol] = append(closedLots, result.NewClosedLots...)
	return nil
}

func handleAssetSplit(s AssetSplit, p *Portfolio) {
	ratio := decimal.NewFromInt32(s.Ratio)
	for _, o := range p.OpenLots[s.Symbol] {
		o.CostBasis = o.CostBasis.Div(ratio)
		o.Quantity = o.Quantity.Mul(ratio)
	}
}

func dateStr(t time.Time) string {
	return t.Format("2006-01-02")
}
