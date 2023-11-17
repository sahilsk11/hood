package service

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	. "hood/internal/domain"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type HoldingsService interface {
	Get(tx *sql.Tx, tradingAccountID uuid.UUID) (*Holdings, error)
}

type holdingsServiceHandler struct{}

func NewHoldingsService() HoldingsService {
	return holdingsServiceHandler{}
}

func (h holdingsServiceHandler) Get(tx *sql.Tx, tradingAccountID uuid.UUID) (*Holdings, error) {
	// fuckkkkk the open lot logic is broken
	// so i dont think this works. iirc the new
	// plaid ingestion code doesn't even try
	// to add open lots
	// .. maybe just nuke the tables and retry
	// this db structure is a huge pain. i think
	// we should tear it all away and re-run
	// all trades every time we want something
	openLots, err := db.GetCurrentOpenLots(tx, tradingAccountID)
	if err != nil {
		return nil, err
	}

	mappedOpenLots := map[string][]*OpenLot{}
	for _, lot := range openLots {
		symbol := lot.GetSymbol()
		if _, ok := mappedOpenLots[symbol]; !ok {
			mappedOpenLots[symbol] = []*OpenLot{}
		}
		mappedOpenLots[symbol] = append(mappedOpenLots[symbol], &lot)
	}

	return Portfolio{
		OpenLots: mappedOpenLots,
		// Cash:     decimal.Zero,
	}.ToHoldings(), nil
}

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
func Playback(in Events) (*HistoricPortfolio, error) {
	hp := HistoricPortfolio([]domain.PortfolioOnDate{})

	events := mergeEvents(in)
	if len(events) == 0 {
		return nil, fmt.Errorf("no events found")
	}

	for _, e := range events {
		portfolio := hp.Latest()
		if portfolio == nil {
			portfolio = &domain.Portfolio{}
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
		hp = append(hp, PortfolioOnDate{
			Portfolio: *portfolio,
			Date:      e.GetDate(),
		})
	}

	return &hp, nil
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
	result, err := PreviewSellOrder(t, openLots)
	if err != nil {
		return err
	}

	p.Cash = p.Cash.Add(result.CashDelta)
	p.OpenLots[t.Symbol] = result.OpenLots
	if len(p.OpenLots[t.Symbol]) == 0 {
		delete(p.OpenLots, t.Symbol)
	}

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

type ProcessSellOrderResult struct {
	NewClosedLots []domain.ClosedLot
	OpenLots      []*domain.OpenLot // current state of open lots
	NewOpenLots   []domain.OpenLot  // lots that were changed
	CashDelta     decimal.Decimal
}

// Selling an asset involves closing currently open lots. In doing this, we may either
// close all open lots for the asset, or close some. The latter requires us to modify
// the existing open lot. Actually, both require us to modify the open lot
//
// This function does the "heavy lifting" to determine which lots should be sold
// without actually selling them. It's only exported because we re-use this logic
// when simulating what a sell order would do
func PreviewSellOrder(t domain.Trade, openLots []*domain.OpenLot) (*ProcessSellOrderResult, error) {
	cashDelta := (t.Price.Mul(t.Quantity))
	closedLots := []domain.ClosedLot{}
	newOpenLots := []domain.OpenLot{}
	// ensure lots are in FIFO
	// could make this dynamic for LIFO systems
	sort.Slice(openLots, func(i, j int) bool {
		return openLots[i].GetPurchaseDate().Before(openLots[j].GetPurchaseDate())
	})

	remainingSellQuantity := t.Quantity
	for remainingSellQuantity.GreaterThan(decimal.Zero) {
		if len(openLots) == 0 {
			return nil, fmt.Errorf("no remaining open lots to execute trade id %d; %f shares outstanding", t.TradeID, remainingSellQuantity.InexactFloat64())
		}
		lot := openLots[0]
		quantitySold := remainingSellQuantity
		if lot.Quantity.LessThan(remainingSellQuantity) {
			quantitySold = lot.Quantity
		}

		remainingSellQuantity = remainingSellQuantity.Sub(quantitySold)
		lot.Quantity = lot.Quantity.Sub(quantitySold)
		lot.OpenLotID = nil // no longer the DB model we're looking at
		lot.Date = t.Date
		newOpenLots = append(newOpenLots, *lot.DeepCopy())
		if lot.Quantity.Equal(decimal.Zero) {
			openLots = openLots[1:]
		}

		gains := (t.Price.Sub(lot.CostBasis)).Mul(quantitySold)
		gainsType := model.GainsType_ShortTerm
		daysBetween := t.Date.Sub(lot.GetPurchaseDate())
		if daysBetween.Hours()/24 >= 365 {
			gainsType = model.GainsType_LongTerm
		}
		closedLots = append(closedLots, domain.ClosedLot{
			OpenLot:       lot,
			SellTrade:     &t,
			Quantity:      quantitySold,
			GainsType:     gainsType,
			RealizedGains: gains,
		})
	}

	return &ProcessSellOrderResult{
		CashDelta:     cashDelta,
		OpenLots:      openLots,
		NewClosedLots: closedLots,
		NewOpenLots:   newOpenLots,
	}, nil
}
