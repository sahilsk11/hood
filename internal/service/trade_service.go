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
	"hood/internal/repository"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// updates open/closed lots in db

type TradeIngestionService interface {
	ProcessBuyOrder(ctx context.Context, tx *sql.Tx, in domain.Trade) (*domain.Trade, *domain.OpenLot, error)
	ProcessSellOrder(ctx context.Context, tx *sql.Tx, in domain.Trade) (*domain.Trade, []*model.ClosedLot, error)
	AddAssetSplit(ctx context.Context, tx *sql.Tx, split model.AssetSplit, tradingAccountID uuid.UUID) (*model.AssetSplit, []model.AppliedAssetSplit, error)

	ProcessTdaBuyOrder(ctx context.Context, tx *sql.Tx, input domain.Trade, tdaTxID *int64) (*domain.Trade, *domain.OpenLot, error)
	ProcessTdaSellOrder(ctx context.Context, tx *sql.Tx, input domain.Trade, tdaTxID *int64) (*domain.Trade, []*model.ClosedLot, error)
}

type tradeIngestionHandler struct {
	TradeRepository repository.TradeRepository
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

	insertedTrades, err := h.TradeRepository.Add(tx, []domain.Trade{t})
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
	openLots, err := db.GetOpenLots(ctx, tx, t.Symbol, t.TradingAccountID)
	if err != nil {
		return nil, nil, err
	}

	insertedTrades, err := h.TradeRepository.Add(tx, []domain.Trade{t})
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

func (h tradeIngestionHandler) AddAssetSplit(ctx context.Context, tx *sql.Tx, split model.AssetSplit, tradingAccountID uuid.UUID) (*model.AssetSplit, []model.AppliedAssetSplit, error) {
	split.CreatedAt = time.Now().UTC()
	insertedSplits, err := db.AddAssetsSplits(tx, []*model.AssetSplit{&split})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedSplits) == 0 {
		return nil, nil, nil
	}
	insertedSplit := insertedSplits[0]
	tdaLots, err := db.GetOpenLots(ctx, tx, split.Symbol, tradingAccountID)
	if err != nil {
		return nil, nil, err
	}
	rhLots, err := db.GetOpenLots(ctx, tx, split.Symbol, tradingAccountID)
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
