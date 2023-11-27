package service

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	. "hood/internal/domain"
	"hood/internal/repository"
	"hood/internal/util"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type HoldingsService interface {
	GetHistoricPortfolio(tx *sql.Tx, tradingAccountID uuid.UUID) (*HistoricPortfolio, error)
	GetCurrentPortfolio(tx *sql.Tx, tradingAccountID uuid.UUID) (*Holdings, error)
}

type holdingsServiceHandler struct {
	TradeRepository          repository.TradeRepository
	TradingAccountRepository repository.TradingAccountRepository
	PositionsRepository      repository.PositionsRepository
}

func NewHoldingsService(
	tradeRepository repository.TradeRepository,
	tradingAccountRepository repository.TradingAccountRepository,
	positionsRepository repository.PositionsRepository,
) HoldingsService {
	return holdingsServiceHandler{
		TradeRepository: tradeRepository,
		// MdsRepository:   mdsRepository,
		TradingAccountRepository: tradingAccountRepository,
		PositionsRepository:      positionsRepository,
	}
}

// should we make this so it only works on trade accounts?
func (h holdingsServiceHandler) GetHistoricPortfolio(tx *sql.Tx, tradingAccountID uuid.UUID) (*HistoricPortfolio, error) {
	tradeSymbols := util.NewSet()

	trades, err := h.TradeRepository.List(tx, tradingAccountID)
	if err != nil {
		return nil, err
	}
	for _, t := range trades {
		tradeSymbols.Add(t.Symbol)
	}

	assetSplits, err := db.GetHistoricAssetSplits(tx)
	if err != nil {
		return nil, err
	}

	// TODO - get this populated for other flows

	tranfers, err := db.GetHistoricTransfers(tx, model.CustodianType_Tda)
	if err != nil {
		return nil, err
	}

	// todo - migrate these queries to repo pattern

	dividends, err := db.GetHistoricDividends(tx, model.CustodianType_Tda)
	if err != nil {
		return nil, err
	}

	// dividends can be retrieved from plaid... i mean we just treat like
	// a transfer. ultimately matters for more accurate cash calculation
	// i think, which should be important for performance only

	events := Events{
		Trades:      trades,
		AssetSplits: assetSplits,
		Transfers:   tranfers,
		Dividends:   dividends,
	}

	historicPortfolio, err := Playback(events)
	if err != nil {
		return nil, err
	}

	return historicPortfolio, nil
}

func (h holdingsServiceHandler) GetCurrentPortfolio(tx *sql.Tx, tradingAccountID uuid.UUID) (*Holdings, error) {
	tradingAccount, err := h.TradingAccountRepository.Get(tx, tradingAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trading account with id %s: %w", tradingAccountID.String(), err)
	}
	if tradingAccount.DataSource == model.TradingAccountDataSourceType_Trades {
		historic, err := h.GetHistoricPortfolio(tx, tradingAccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get historic portfolio: %w", err)
		}
		return historic.Latest().ToHoldings(), nil
	}

	positions, err := h.PositionsRepository.List(tx, tradingAccountID)
	if err != nil {
		return nil, err
	}
	kys := map[string]*Position{}
	for _, p := range positions {
		t := p // kys
		kys[p.Symbol] = &t
	}

	// todo - handle cash
	return &domain.Holdings{
		Positions: kys,
		Cash:      decimal.Zero, // fuck idk what to do here
	}, nil
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
	Transfers   []Transfer // make sure plaid is tracking this
	Dividends   []Dividend // hmm..
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
