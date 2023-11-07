package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"sort"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/shopspring/decimal"
)

// consider changing these names from "Add" to something better

type TradeIngestionService interface {
	ProcessBuyOrder(ctx context.Context, tx *sql.Tx, in domain.Trade) (*domain.Trade, *domain.OpenLot, error)
	ProcessSellOrder(ctx context.Context, tx *sql.Tx, in domain.Trade) (*domain.Trade, []*model.ClosedLot, error)
	AddAssetSplit(ctx context.Context, tx *sql.Tx, split model.AssetSplit) (*model.AssetSplit, []model.AppliedAssetSplit, error)

	ProcessTdaBuyOrder(ctx context.Context, tx *sql.Tx, input domain.Trade, tdaTxID *int64) (*domain.Trade, *domain.OpenLot, error)
	ProcessTdaSellOrder(ctx context.Context, tx *sql.Tx, input domain.Trade, tdaTxID *int64) (*domain.Trade, []*model.ClosedLot, error)
}

type tradeIngestionHandler struct {
}

func NewTradeIngestionService() TradeIngestionService {
	return tradeIngestionHandler{}
}

func (h tradeIngestionHandler) ProcessTdaBuyOrder(ctx context.Context, tx *sql.Tx, t domain.Trade, tdaTransactionID *int64) (*domain.Trade, *domain.OpenLot, error) {
	trade, lots, err := h.ProcessBuyOrder(ctx, tx, t)
	if err != nil {
		return nil, nil, err
	}

	tdaOrder := model.TdaTrade{
		TdaTransactionID: tdaTransactionID,
		TradeID:          *trade.TradeID,
	}

	err = db.AddTdaTrade(tx, tdaOrder)
	if err != nil {
		return nil, nil, err
	}

	return trade, lots, nil
}

func (h tradeIngestionHandler) ProcessTdaSellOrder(ctx context.Context, tx *sql.Tx, t domain.Trade, tdaTransactionID *int64) (*domain.Trade, []*model.ClosedLot, error) {
	trade, lots, err := h.ProcessSellOrder(ctx, tx, t)
	if err != nil {
		return nil, nil, err
	}

	tdaOrder := model.TdaTrade{
		TdaTransactionID: tdaTransactionID,
		TradeID:          *trade.TradeID,
	}

	err = db.AddTdaTrade(tx, tdaOrder)
	if err != nil {
		return nil, nil, err
	}

	return trade, lots, nil
}

func (h tradeIngestionHandler) ProcessBuyOrder(ctx context.Context, tx *sql.Tx, t domain.Trade) (*domain.Trade, *domain.OpenLot, error) {
	if t.Action != model.TradeActionType_Buy {
		return nil, nil, fmt.Errorf("failed to process buy order with action %s", t.Action.String())
	}

	insertedTrades, err := db.AddTrades(ctx, tx, []domain.Trade{t})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add trades for buy order: %w", err)
	}
	if len(insertedTrades) == 0 {
		return nil, nil, nil
	}

	insertedTrade := insertedTrades[0]
	newLot := domain.OpenLot{
		CostBasis: insertedTrade.Price,
		Quantity:  insertedTrade.Quantity,
		Trade:     &insertedTrade,
	}
	insertedLots, err := db.AddOpenLots(ctx, tx, []domain.OpenLot{newLot})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, errors.New("no inserted open lots")
	}
	insertedLot := insertedLots[0]

	return &insertedTrade, &insertedLot, nil
}

func (h tradeIngestionHandler) ProcessSellOrder(ctx context.Context, tx *sql.Tx, t domain.Trade) (*domain.Trade, []*model.ClosedLot, error) {
	if t.Action != model.TradeActionType_Sell {
		return nil, nil, fmt.Errorf("failed to process sell order with action %s", t.Action.String())
	}
	openLots, err := db.GetOpenLots(ctx, tx, t.Symbol, t.Custodian)
	if err != nil {
		return nil, nil, err
	}

	insertedTrades, err := db.AddTrades(ctx, tx, []domain.Trade{t})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, nil
	}
	t = insertedTrades[0]
	sellOrderResult, err := PreviewSellOrder(t, domain.OpenLots(openLots).Ptr())
	if err != nil {
		return nil, nil, err
	}
	// TODO - modify open lots
	insertedClosedLots, err := db.AddClosedLots(ctx, tx, sellOrderResult.NewClosedLots)
	if err != nil {
		return nil, nil, err
	}

	return &t, insertedClosedLots, nil
}

func validateTrade(t domain.Trade) error {
	if t.Quantity.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("trade must have quantity higher than 0, received %f", t.Quantity.InexactFloat64())
	}
	if len(t.Symbol) == 0 {
		return errors.New("trade has invalid ticker (empty string)")
	}
	if t.Quantity.LessThan(decimal.Zero) {
		return fmt.Errorf("trade has invalid cost basi %f", t.Price.InexactFloat64())
	}

	return nil
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

func (h tradeIngestionHandler) AddAssetSplit(ctx context.Context, tx *sql.Tx, split model.AssetSplit) (*model.AssetSplit, []model.AppliedAssetSplit, error) {
	split.CreatedAt = time.Now().UTC()
	insertedSplits, err := db.AddAssetsSplits(tx, []*model.AssetSplit{&split})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedSplits) == 0 {
		return nil, nil, nil
	}
	insertedSplit := insertedSplits[0]
	tdaLots, err := db.GetOpenLots(ctx, tx, split.Symbol, model.CustodianType_Tda)
	if err != nil {
		return nil, nil, err
	}
	rhLots, err := db.GetOpenLots(ctx, tx, split.Symbol, model.CustodianType_Robinhood)
	if err != nil {
		return nil, nil, err
	}
	lots := append(tdaLots, rhLots...)

	ratio := decimal.NewFromInt32(insertedSplit.Ratio)
	appliedSplits := []model.AppliedAssetSplit{}
	for _, lot := range lots {
		dbLot := model.OpenLot{
			OpenLotID: *lot.OpenLotID,
			CostBasis: lot.CostBasis.Div(ratio),
			Quantity:  lot.Quantity.Mul(ratio),
		}
		columnList := postgres.ColumnList{table.OpenLot.CostBasis, table.OpenLot.Quantity}
		updatedOpenLot, err := db.UpdateOpenLotInDb(ctx, tx, dbLot, columnList)
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
	insertedAppliedSplits, err := db.AddAppliedAssetSplits(ctx, tx, appliedSplits)
	if err != nil {
		return nil, nil, err
	}

	return &insertedSplit, insertedAppliedSplits, nil
}
