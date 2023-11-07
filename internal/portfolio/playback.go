package portfolio

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	. "hood/internal/domain"
	"hood/internal/service"
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

// given historical events, calculate all portfolios
func Playback(initialPortfolio *Portfolio, in Events) (*HistoricPortfolio, error) {
	hp := NewHistoricPortfolio(nil)
	if initialPortfolio != nil {
		hp.Append(*initialPortfolio)
	}

	events := mergeEvents(in)
	if len(events) == 0 {
		return nil, fmt.Errorf("no events found")
	}

	for _, e := range events {
		portfolio := hp.Latest()
		if portfolio == nil {
			portfolio = NewEmptyPortfolio()
		}
		portfolio = portfolio.DeepCopy()
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
		portfolio.LastAction = e.GetDate()
		hp.Append(*portfolio)
	}
	return hp, nil
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
	result, err := service.PreviewSellOrder(t, openLots)
	if err != nil {
		return err
	}

	p.Cash = p.Cash.Add(result.CashDelta)
	p.OpenLots[t.Symbol] = result.OpenLots
	if len(p.OpenLots[t.Symbol]) == 0 {
		delete(p.OpenLots, t.Symbol)
	}
	p.NewOpenLots = append(p.NewOpenLots, result.NewOpenLots...)

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

func insertPortfolio(tx *sql.Tx, portfolio domain.Portfolio) error {
	cash := portfolio.Cash
	err := db.AddCash(tx, model.Cash{
		Amount:    cash,
		Custodian: model.CustodianType_Robinhood,
		Date:      portfolio.LastAction,
	})
	if err != nil {
		return err
	}
	openLots := []domain.OpenLot{}
	for _, lots := range portfolio.OpenLots {
		for _, lot := range lots {
			openLots = append(openLots, *lot)
		}
	}
	for _, lot := range portfolio.NewOpenLots {
		openLots = append(openLots, lot)
	}
	err = db.AddImmutableOpenLots(tx, openLots)
	if err != nil {
		return err
	}
	return nil
}
