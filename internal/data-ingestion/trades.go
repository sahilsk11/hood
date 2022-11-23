package data_ingestion

import (
	"context"
	"errors"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/util"
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

func AddBuyOrder(ctx context.Context, newTrade model.Trade) (*model.Trade, *model.OpenLot, error) {
	insertedTrades, err := AddTradesToDb(ctx, []*model.Trade{&newTrade})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, errors.New("no inserted trades")
	}
	insertedTrade := insertedTrades[0]
	newLot := model.OpenLot{
		CostBasis:  insertedTrade.CostBasis,
		Quantity:   insertedTrade.Quantity,
		TradeID:    insertedTrade.TradeID,
		CreatedAt:  util.TimePtr(time.Now().UTC()),
		ModifiedAt: util.TimePtr(time.Now().UTC()),
	}
	insertedLots, err := AddOpenLotsToDb(ctx, []*model.OpenLot{&newLot})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, errors.New("no inserted open lots")
	}
	insertedLot := insertedLots[0]

	return &insertedTrade, &insertedLot, nil
}

func AddSellOrder(ctx context.Context, newTrade model.Trade) (*model.Trade, *model.OpenLot, error) {
	insertedTrades, err := AddTradesToDb(ctx, []*model.Trade{&newTrade})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, errors.New("no inserted trades")
	}
	insertedTrade := insertedTrades[0]
	newLot := model.OpenLot{
		CostBasis:  insertedTrade.CostBasis,
		Quantity:   insertedTrade.Quantity,
		TradeID:    insertedTrade.TradeID,
		CreatedAt:  util.TimePtr(time.Now().UTC()),
		ModifiedAt: util.TimePtr(time.Now().UTC()),
	}
	insertedLots, err := AddOpenLotsToDb(ctx, []*model.OpenLot{&newLot})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, errors.New("no inserted open lots")
	}
	insertedLot := insertedLots[0]

	return &insertedTrade, &insertedLot, nil
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
	// used for updating caller's list
	// useful for running in a list
	RemainingOpenLots []*model.OpenLot
	NewClosedLots     []*model.ClosedLot
	// Lots that need to be updated in DB
	// if trade is executed
	UpdatedOpenLots []*model.OpenLot
}

func ProcessSellOrder(t model.Trade, openLots []*model.OpenLot) (*ProcessSellOrderResult, error) {
	if err := validateTrade(t); err != nil {
		return nil, err
	}

	// ensure lots are in FIFO
	// could make this dynamic for LIFO systems
	sort.Slice(openLots, func(i, j int) bool {
		return openLots[i].TradeID < openLots[j].TradeID
	})

	remainingSellQuantity := t.Quantity
	updatedOpenLots := []*model.OpenLot{}
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
			CreatedAt:     util.TimePtr(time.Now().UTC()),
			ModifiedAt:    util.TimePtr(time.Now().UTC()),
			RealizedGains: gains,
			GainsType:     gainsType,
		}
		newClosedLots = append(newClosedLots, &newClosedLot)

		lot.Quantity = lot.Quantity.Sub(quantitySold)
		if lot.Quantity.Equal(decimal.Zero) {
			lot.DeletedAt = util.TimePtr(time.Now().UTC())
			openLots = openLots[1:]
		}
		updatedOpenLots = append(updatedOpenLots, lot)

		remainingSellQuantity = remainingSellQuantity.Sub(quantitySold)
	}

	return &ProcessSellOrderResult{
		RemainingOpenLots: openLots,
		NewClosedLots:     newClosedLots,
		UpdatedOpenLots:   updatedOpenLots,
	}, nil
}
