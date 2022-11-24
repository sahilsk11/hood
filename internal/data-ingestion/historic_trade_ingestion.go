package data_ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hood/internal/db/models/postgres/public/model"

	"os"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

type entry struct {
	// tradeEntry and assetSplitEntry
	Asset  string `json:"asset"`
	Action string `json:"action"`
	Date   string `json:"date"`

	// for tradeEntry
	LongDate  string          `json:"long_date"`
	Quantity  decimal.Decimal `json:"quantity"`
	CostBasis decimal.Decimal `json:"cost_basis"`

	// for assetSplitEntry
	Ratio int32 `json:"ratio"`
}

type entryIterator struct {
	tradeEntries      []*model.Trade
	assetSplitEntries []*model.AssetSplit
}

func newEntryiterator(tradeEntries []*model.Trade, assetSplitEntries []*model.AssetSplit) entryIterator {
	sort.Slice(tradeEntries, func(i, j int) bool {
		return tradeEntries[i].Date.Unix() < tradeEntries[j].Date.Unix()
	})
	sort.Slice(assetSplitEntries, func(i, j int) bool {
		return assetSplitEntries[i].Date.Unix() < assetSplitEntries[j].Date.Unix()
	})
	return entryIterator{
		tradeEntries:      tradeEntries,
		assetSplitEntries: assetSplitEntries,
	}
}

func (m entryIterator) hasNext() bool {
	return len(m.tradeEntries)+len(m.assetSplitEntries) > 0
}

// returns the next parsed entry from the "entry list"
// to help avoid the use of interface{}, at most one
// non-nil value will be returned which represents the
// next entry
func (m *entryIterator) next() (*model.Trade, *model.AssetSplit) {
	if len(m.tradeEntries) == 0 && len(m.assetSplitEntries) == 0 {
		return nil, nil
	}
	if len(m.tradeEntries) == 0 {
		nextAssetSplit := m.assetSplitEntries[0]
		m.assetSplitEntries = m.assetSplitEntries[1:]
		return nil, nextAssetSplit
	}
	if len(m.assetSplitEntries) == 0 {
		nextTrade := m.tradeEntries[0]
		m.tradeEntries = m.tradeEntries[1:]
		return nextTrade, nil
	}

	nextTrade := m.tradeEntries[0]
	nextAssetSplit := m.assetSplitEntries[0]
	if nextTrade.Date.Unix() > nextAssetSplit.Date.Unix() {
		m.assetSplitEntries = m.assetSplitEntries[1:]
		return nil, nextAssetSplit
	}
	m.tradeEntries = m.tradeEntries[1:]
	return nextTrade, nil
}

type OutfileEntries struct {
	Trades      []*model.Trade
	AssetSplits []*model.AssetSplit
}

// ParseFromOutfile reads the output JSON generated
// by rh.py
func ParseEntriesFromOutfile() (*OutfileEntries, error) {
	f, err := os.ReadFile("new.json")
	if err != nil {
		return nil, fmt.Errorf("could not open out.json: %w", err)
	}

	trades := []*model.Trade{}
	splits := []*model.AssetSplit{}

	var entries []entry
	err = json.Unmarshal(f, &entries)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal file into entries: %w", err)
	}

	for _, entry := range entries {
		if entry.Action == "BUY" || entry.Action == "SELL" {
			trade, err := parseTradeFromEntry(entry)
			if err != nil {
				entryStr, _ := json.Marshal(entry)
				return nil, fmt.Errorf("failed to parse trade from entry: %s, err: %w", string(entryStr), err)
			}
			trades = append(trades, trade)
		} else if entry.Action == "SPLIT" {
			split, err := parseSplitFromEntry(entry)
			if err != nil {
				entryStr, _ := json.Marshal(entry)
				return nil, fmt.Errorf("failed to parse asset split from entry: %s, err: %w", string(entryStr), err)
			}
			splits = append(splits, split)
		} else {
			entryStr, _ := json.Marshal(entry)
			return nil, fmt.Errorf("unknown action '%s' in entry %s", entry.Action, string(entryStr))
		}
	}

	return &OutfileEntries{
		Trades:      trades,
		AssetSplits: splits,
	}, nil
}

func parseTradeFromEntry(e entry) (*model.Trade, error) {
	ticker, err := nameToTicker(e.Asset)
	if err != nil {
		return nil, err
	}
	var action model.TradeActionType
	if e.Action == "BUY" {
		action = model.TradeActionType_Buy
	} else if e.Action == "SELL" {
		action = model.TradeActionType_Sell
	}
	if e.LongDate == "" {
		return nil, errors.New("trade missing long date field")
	}
	date, err := time.Parse("Jan 2, 2006 3:04 PM MST", e.LongDate)
	if err != nil {
		return nil, err
	}

	t := model.Trade{
		Symbol:    ticker,
		Action:    action,
		CostBasis: e.CostBasis,
		Quantity:  e.Quantity,
		Date:      date,
	}

	return &t, nil
}

func parseSplitFromEntry(e entry) (*model.AssetSplit, error) {
	ticker, err := nameToTicker(e.Asset)
	if err != nil {
		return nil, err
	}
	date, err := time.Parse("2006-01-02", e.Date)
	if err != nil {
		return nil, err
	}
	ratio := e.Ratio
	if ratio <= 0 {
		return nil, fmt.Errorf("invalid asset split ratio %d", ratio)
	}
	return &model.AssetSplit{
		Symbol: ticker,
		Ratio:  ratio,
		Date:   date,
	}, nil
}

func ProcessHistoricTrades(ctx context.Context, i entryIterator) error {
	for i.hasNext() {
		nextTrade, nextSplit := i.next()
		if nextSplit != nil {
			_, _, err := AddAssetSplit(ctx, *nextSplit)
			if err != nil {
				return fmt.Errorf("failed to add asset split: %w", err)
			}
		} else if nextTrade != nil {
			if nextTrade.Action == model.TradeActionType_Buy {
				_, _, err := AddBuyOrder(ctx, *nextTrade)
				if err != nil {
					return fmt.Errorf("failed to add buy order %v: %w", *nextTrade, err)
				}
			} else {
				_, _, err := AddSellOrder(ctx, *nextTrade)
				if err != nil {
					return fmt.Errorf("failed to add sell order %v: %w", *nextTrade, err)
				}
			}
		}
	}

	return nil
}

func ProcessOutfile(ctx context.Context) error {
	outfileEntries, err := ParseEntriesFromOutfile()
	if err != nil {
		return err
	}
	entryIterator := newEntryiterator(outfileEntries.Trades, outfileEntries.AssetSplits)
	err = ProcessHistoricTrades(ctx, entryIterator)
	if err != nil {
		return err
	}

	return nil
}
