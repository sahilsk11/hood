package trade_ingestion

import (
	"context"
	"errors"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"math"
	"sort"
	"time"

	"hood/internal/util"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/shopspring/decimal"
)

func AddBuyOrder(ctx context.Context, newTrade model.Trade) (*model.Trade, *model.OpenLot, error) {
	newTrade.CreatedAt = time.Now().UTC()
	newTrade.ModifiedAt = time.Now().UTC()
	insertedTrades, err := db.AddTradesToDb(ctx, []*model.Trade{&newTrade})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, nil
	}
	insertedTrade := insertedTrades[0]
	newLot := model.OpenLot{
		CostBasis:  insertedTrade.CostBasis,
		Quantity:   insertedTrade.Quantity,
		TradeID:    insertedTrade.TradeID,
		CreatedAt:  time.Now().UTC(),
		ModifiedAt: time.Now().UTC(),
	}
	insertedLots, err := db.AddOpenLotsToDb(ctx, []*model.OpenLot{&newLot})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, errors.New("no inserted open lots")
	}
	insertedLot := insertedLots[0]

	return &insertedTrade, &insertedLot, nil
}

func AddSellOrder(ctx context.Context, newTrade model.Trade) (*model.Trade, []*model.ClosedLot, error) {
	openLots, err := db.GetOpenLots(ctx, newTrade.Symbol)
	if err != nil {
		return nil, nil, err
	}
	newTrade.CreatedAt = time.Now().UTC()
	newTrade.ModifiedAt = time.Now().UTC()
	insertedTrades, err := db.AddTradesToDb(ctx, []*model.Trade{&newTrade})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, nil
	}
	insertedTrade := insertedTrades[0]
	sellOrderResult, err := ProcessSellOrder(insertedTrade, openLots)
	if err != nil {
		return nil, nil, err
	}
	insertedClosedLots, err := db.AddClosedLotsToDb(ctx, sellOrderResult.NewClosedLots)
	if err != nil {
		return nil, nil, err
	}
	for _, updatedOpenLot := range sellOrderResult.UpdatedOpenLots {
		columnlist := postgres.ColumnList{table.OpenLot.Quantity}
		dbOpenLot := model.OpenLot{
			OpenLotID: updatedOpenLot.OpenLotID,
			Quantity:  updatedOpenLot.Quantity,
		}
		if updatedOpenLot.Quantity.Equal(decimal.Zero) {
			columnlist = append(columnlist, table.OpenLot.DeletedAt)
			dbOpenLot.DeletedAt = util.TimePtr(time.Now().UTC())
		}
		_, err = db.UpdateOpenLotInDb(ctx, dbOpenLot, columnlist)
		if err != nil {
			return nil, nil, err
		}
	}

	return &insertedTrade, insertedClosedLots, nil
}

func validateTrade(t model.Trade) error {
	if t.Quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("trade must have quantity higher than 0, received %f", t.Quantity.InexactFloat64())
	}
	if len(t.Symbol) == 0 {
		return errors.New("trade has invalid ticker (empty string)")
	}
	if t.Quantity.LessThan(decimal.Zero) {
		return fmt.Errorf("trade has invalid cost basi %f", t.CostBasis.InexactFloat64())
	}

	return nil
}

type ProcessSellOrderResult struct {
	NewClosedLots []*model.ClosedLot
	// Lots that need to be updated in DB
	// if trade is executed
	UpdatedOpenLots []*domain.OpenLot
}

func ProcessSellOrder(t model.Trade, openLots []*domain.OpenLot) (*ProcessSellOrderResult, error) {
	if err := validateTrade(t); err != nil {
		return nil, err
	}

	// ensure lots are in FIFO
	// could make this dynamic for LIFO systems
	sort.Slice(openLots, func(i, j int) bool {
		return openLots[i].PurchaseDate.Unix() < openLots[j].PurchaseDate.Unix()
	})

	remainingSellQuantity := t.Quantity
	updatedOpenLots := []*domain.OpenLot{}
	newClosedLots := []*model.ClosedLot{}

	for remainingSellQuantity.GreaterThan(decimal.Zero) {
		if len(openLots) == 0 {
			return nil, fmt.Errorf("no remaining open lots to execute trade id %d; %f shares outstanding", t.TradeID, remainingSellQuantity.InexactFloat64())
		}
		lot := openLots[0]
		quantitySold := remainingSellQuantity
		if lot.Quantity.LessThan(remainingSellQuantity) {
			quantitySold = lot.Quantity
		}

		gains := (t.CostBasis.Sub(lot.CostBasis)).Mul(quantitySold)

		daysBetween := math.Abs(float64(time.Until(t.Date).Hours() / 24))
		gainsType := model.GainsType_ShortTerm
		if daysBetween >= 365 {
			gainsType = model.GainsType_LongTerm
		}

		newClosedLot := model.ClosedLot{
			BuyTradeID:    lot.TradeID,
			SellTradeID:   t.TradeID,
			Quantity:      quantitySold,
			CreatedAt:     time.Now().UTC(),
			ModifiedAt:    time.Now().UTC(),
			RealizedGains: gains,
			GainsType:     gainsType,
		}
		newClosedLots = append(newClosedLots, &newClosedLot)

		lot.Quantity = lot.Quantity.Sub(quantitySold)
		if lot.Quantity.Equal(decimal.Zero) {
			openLots = openLots[1:]
		}
		updatedOpenLots = append(updatedOpenLots, lot)

		remainingSellQuantity = remainingSellQuantity.Sub(quantitySold)
	}

	return &ProcessSellOrderResult{
		NewClosedLots:   newClosedLots,
		UpdatedOpenLots: updatedOpenLots,
	}, nil
}

func AddAssetSplit(ctx context.Context, split model.AssetSplit) (*model.AssetSplit, []model.AppliedAssetSplit, error) {
	split.CreatedAt = time.Now().UTC()
	insertedSplits, err := db.AddAssetsSplitsToDb(ctx, []*model.AssetSplit{&split})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedSplits) == 0 {
		return nil, nil, nil
	}
	insertedSplit := insertedSplits[0]
	lots, err := db.GetOpenLots(ctx, split.Symbol)
	if err != nil {
		return nil, nil, err
	}

	ratio := decimal.NewFromInt32(insertedSplit.Ratio)
	appliedSplits := []model.AppliedAssetSplit{}
	for _, lot := range lots {
		dbLot := model.OpenLot{
			OpenLotID: lot.OpenLotID,
			CostBasis: lot.CostBasis.Div(ratio),
			Quantity:  lot.Quantity.Mul(ratio),
		}
		columnList := postgres.ColumnList{table.OpenLot.CostBasis, table.OpenLot.Quantity}
		updatedOpenLot, err := db.UpdateOpenLotInDb(ctx, dbLot, columnList)
		if err != nil {
			return nil, nil, err
		}
		appliedSplit := model.AppliedAssetSplit{
			AssetSplitID: insertedSplit.AssetSplitID,
			OpenLotID:    updatedOpenLot.OpenLotID,
			AppliedAt:    time.Now().UTC(),
		}
		appliedSplits = append(appliedSplits, appliedSplit)
	}
	insertedAppliedSplits, err := db.AddAppliedAssetSplitsToDb(ctx, appliedSplits)
	if err != nil {
		return nil, nil, err
	}

	return &insertedSplit, insertedAppliedSplits, nil
}
