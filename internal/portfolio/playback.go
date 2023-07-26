package portfolio

import (
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/domain"
	"hood/internal/trade"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// replay historic events

func mergeEvents(
	trades []Trade,
	assetSplits []AssetSplit,
	transfers []Transfer,
) []TradeEvent {
	out := []TradeEvent{}
	for _, t := range trades {
		out = append(out, t)
	}
	for _, t := range assetSplits {
		out = append(out, t)
	}
	for _, t := range transfers {
		out = append(out, t)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].GetDate().After(out[j].GetDate())
	})
	return out
}

// given new data, figure out what to do
// should be dry. can have another func
// for committing
func Playback(
	trades []Trade,
	assetSplits []AssetSplit,
	transfers []Transfer,
) error {
	portfolio := &Portfolio{
		OpenLots:   map[string][]OpenLot{},
		ClosedLots: map[string][]ClosedLot{},
	}

	events := mergeEvents(trades, assetSplits, transfers)

	for _, e := range events {
		switch e.(type) {
		case Trade:
			t := e.(Trade)
			if t.Action == model.TradeActionType_Buy {
				handleBuy(t, portfolio)
			} else if t.Action == model.TradeActionType_Sell {
				handleSell(t, portfolio)
			}
		case AssetSplit:
			handleAssetSplit(e.(AssetSplit), portfolio)
		case Transfer:
			t := e.(Transfer)
			portfolio.Cash = portfolio.Cash.Add(t.Amount)
		}

	}
	return nil
}

func handleBuy(t Trade, p *Portfolio) {
	if _, ok := p.OpenLots[t.Symbol]; !ok {
		p.OpenLots[t.Symbol] = []OpenLot{}
	}
	newLot := OpenLot{
		LotID:     uuid.New(),
		Trade:     &t,
		Quantity:  t.Quantity,
		CostBasis: t.Price,
		Date:      t.Date,
	}
	p.OpenLots[t.Symbol] = append(p.OpenLots[t.Symbol], newLot)
	p.Cash = p.Cash.Sub(t.Price.Mul(t.Quantity))
}

func handleSell(t Trade, p *Portfolio) error {
	openLots := []OpenLot{}
	if lots, ok := p.OpenLots[t.Symbol]; !ok {
		openLots = lots
	}
	closedLots := []ClosedLot{}
	if lots, ok := p.ClosedLots[t.Symbol]; !ok {
		closedLots = lots
	}
	result, err := trade.PreviewSellOrder(t, openLots)
	if err != nil {
		return err
	}
	closedLots = append(closedLots, result.NewClosedLots...)

	p.Cash = p.Cash.Add(result.CashDelta)
	p.ClosedLots[t.Symbol] = closedLots

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
