package data_ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/util"

	"log"
	"os"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

type outEntry struct {
	Date        string          `json:"date"`
	Asset       string          `json:"asset"`
	Action      string          `json:"action"`
	Quantity    decimal.Decimal `json:"quantity"`
	CostBasis   decimal.Decimal `json:"cost_basis"`
	Description string          `json:"description"`
	Ratio       int             `json:"ratio"`
	LongDate    string          `json:"long_date"`
}

// ParseFromOutfile reads the output JSON generated
// by rh.py
func ParseEntriesFromOutfile() ([]*model.Trade, []*model.AssetSplit, error) {
	f, err := os.ReadFile("out.json")
	if err != nil {
		return nil, nil, fmt.Errorf("could not open out.json: %w", err)
	}

	trades := []*model.Trade{}
	splits := []*model.AssetSplit{}

	reverse := func(numbers []*outEntry) []*outEntry {
		for i := 0; i < len(numbers)/2; i++ {
			j := len(numbers) - i - 1
			numbers[i], numbers[j] = numbers[j], numbers[i]
		}
		return numbers
	}

	var fileContents []*outEntry
	err = json.Unmarshal(f, &fileContents)
	if err != nil {
		return nil, nil, fmt.Errorf("could not unmarshal file into outEntry: %w", err)
	}

	fileContents = reverse(fileContents)

	for _, entry := range fileContents {
		if entry.Action == "BUY" || entry.Action == "SELL" {
			t, err := time.Parse("2006-01-02", entry.Date)
			if err != nil {
				return nil, nil, fmt.Errorf("could not parse trade date: %w", err)
			}
			name := entry.Asset
			ticker, err := nameToTicker(name)
			if err != nil {
				return nil, nil, err
			}
			trade := model.Trade{
				Symbol:      ticker,
				Action:      model.TradeActionType(entry.Action),
				Quantity:    entry.Quantity,
				CostBasis:   entry.CostBasis,
				Date:        t,
				Description: nil,
			}
			trades = append(trades, &trade)
		} else if entry.Action == "SPLIT" {
			t, err := time.Parse("2006-01-02", entry.Date)
			if err != nil {
				return nil, nil, fmt.Errorf("could not parse split date: %w", err)
			}

			ratio := entry.Ratio
			ticker, err := nameToTicker(entry.Asset)
			if err != nil {
				return nil, nil, err
			}
			split := model.AssetSplit{
				Symbol:    ticker,
				Ratio:     int32(ratio),
				Date:      t,
				CreatedAt: util.TimePtr(time.Now().UTC()),
			}
			splits = append(splits, &split)
		}
	}

	return trades, splits, nil
}

func findNextTrade(symbol string, remainingTrades []model.Trade) *model.Trade {
	for _, t := range remainingTrades {
		if t.Symbol == symbol {
			return &t
		}
	}
	return nil
}

func findNextSplit(tradeDate time.Time, splits []model.AssetSplit) *model.AssetSplit {
	for len(splits) > 0 && tradeDate.After(splits[0].Date) {
		splits = splits[1:]
	}
	if len(splits) == 0 {
		return nil
	}
	return &splits[0]
}

func shouldApplySplit(upcomingSplit model.AssetSplit, nextTrade *model.Trade) bool {
	if nextTrade == nil {
		return true
	}
	if upcomingSplit.Date.Before(nextTrade.Date) {
		return true
	}
	return false
}

type ProcessTradesOutput struct {
	OpenLots              []*model.OpenLot
	ClosedLots            []*model.ClosedLot
	AppliedAssetSplitsMap map[int32]model.AppliedAssetSplit
}

func ProcessHistoricTrades(trades []model.Trade, splits []model.AssetSplit) (*ProcessTradesOutput, error) {
	openLotsMap := make(map[string][]*model.OpenLot)
	deletedOpenLots := []*model.OpenLot{}
	closedLots := []*model.ClosedLot{}
	// keyed by trade id because we don't know lot_id
	// until after insertion
	appliedAssetSplitsMap := map[int32]model.AppliedAssetSplit{}

	// sequentially sort splits
	sort.Slice(splits, func(i, j int) bool {
		return splits[i].Date.Unix() < splits[j].Date.Unix()
	})
	assetSplitMap := map[string][]model.AssetSplit{}
	for _, split := range splits {
		if _, ok := assetSplitMap[split.Symbol]; !ok {
			assetSplitMap[split.Symbol] = []model.AssetSplit{}
		}
		assetSplitMap[split.Symbol] = append(assetSplitMap[split.Symbol], split)
	}

	sort.Slice(trades, func(i, j int) bool {
		return trades[i].TradeID < trades[j].TradeID
	})
	sort.SliceStable(trades, func(i, j int) bool {
		return trades[i].Date.Unix() < trades[j].Date.Unix()
	})

	for i, t := range trades {
		if t.Action == model.TradeActionType_Buy {
			var assetSplit *model.AssetSplit
			// decide if there is an upcoming split
			if upcomingSplit := findNextSplit(t.Date, assetSplitMap[t.Symbol]); upcomingSplit != nil {
				nextTrade := findNextTrade(t.Symbol, trades[i+1:])
				if shouldApplySplit(*upcomingSplit, nextTrade) {
					assetSplit = upcomingSplit
				}
			}

			newLot := model.OpenLot{
				TradeID:    t.TradeID,
				CostBasis:  t.CostBasis,
				Quantity:   t.Quantity,
				CreatedAt:  time.Now().UTC(),
				ModifiedAt: time.Now().UTC(),
			}
			if _, ok := openLotsMap[t.Symbol]; !ok {
				openLotsMap[t.Symbol] = []*model.OpenLot{}
			}
			openLotsMap[t.Symbol] = append(openLotsMap[t.Symbol], &newLot)

			// apply asset split to all assets
			if assetSplit != nil {
				assetSplitMap[t.Symbol] = assetSplitMap[t.Symbol][1:]
				for _, lot := range openLotsMap[t.Symbol] {
					ratio := decimal.NewFromInt32(assetSplit.Ratio)
					lot.Quantity = lot.Quantity.Mul(ratio)
					lot.CostBasis = lot.CostBasis.Div(ratio)
					appliedAssetSplitsMap[lot.TradeID] = model.AppliedAssetSplit{
						AssetSplitID: assetSplit.AssetSplitID,
						AppliedAt:    time.Now().UTC(),
					}
				}
			}
		} else {
			sellResult, err := ProcessSellOrder(t, openLotsMap[t.Symbol])
			if err != nil {
				return nil, err
			}
			openLotsMap[t.Symbol] = sellResult.RemainingOpenLots
			// need to store discarded lots because the trading function
			// internally updates the map, but will not include deleted lots
			closedLots = append(closedLots, sellResult.NewClosedLots...)
			for _, lot := range sellResult.UpdatedOpenLots {
				if lot.DeletedAt != nil {
					deletedOpenLots = append(deletedOpenLots, lot)
				}
			}
		}
	}

	openLots := []*model.OpenLot{}
	for _, v := range openLotsMap {
		openLots = append(openLots, v...)
	}

	// need to save deleted lots because asset split application
	// refer to them
	openLots = append(openLots, deletedOpenLots...)

	return &ProcessTradesOutput{
		OpenLots:              openLots,
		ClosedLots:            closedLots,
		AppliedAssetSplitsMap: appliedAssetSplitsMap,
	}, nil
}

func ProcessOutfile(ctx context.Context) error {
	trades, splits, err := ParseEntriesFromOutfile()
	if err != nil {
		log.Fatal(err)
	}

	insertedTrades, err := AddTradesToDb(ctx, trades)
	if err != nil {
		return err
	}

	insertedSplits, err := AddAssetsSplitsToDb(ctx, splits)
	if err != nil {
		return err
	}

	processTradesResponse, err := ProcessHistoricTrades(insertedTrades, insertedSplits)
	if err != nil {
		return err
	}

	insertedLots, err := AddOpenLotsToDb(ctx, processTradesResponse.OpenLots)
	if err != nil {
		return fmt.Errorf("could not add open lots to db: %w", err)
	}

	appliedAssetSplits := []model.AppliedAssetSplit{}
	for _, lot := range insertedLots {
		if appliedSplit, ok := processTradesResponse.AppliedAssetSplitsMap[lot.TradeID]; ok {
			appliedSplit.OpenLotID = lot.OpenLotID
			appliedAssetSplits = append(appliedAssetSplits, appliedSplit)
		}
	}

	_, err = AddAppliedAssetSplitsToDb(ctx, appliedAssetSplits)
	if err != nil {
		return fmt.Errorf("could not add applied asset splits to db: %w", err)
	}

	_, err = AddClosedLotsToDb(ctx, processTradesResponse.ClosedLots)
	if err != nil {
		return fmt.Errorf("could not add closed lots to db: %w", err)
	}

	return nil
}
