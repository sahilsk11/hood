package trade

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	hood_errors "hood/internal"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/db/models/postgres/public/table"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"sort"
	"strings"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/shopspring/decimal"
)

// consider changing these names from "Add" to something better

type TradeIngestionService interface {
	ProcessBuyOrder(ctx context.Context, tx *sql.Tx, in ProcessBuyOrderInput) (*model.Trade, *model.OpenLot, error)
	ProcessSellOrder(ctx context.Context, tx *sql.Tx, in ProcessSellOrderInput) (*model.Trade, []*model.ClosedLot, error)
	AddAssetSplit(ctx context.Context, tx *sql.Tx, split model.AssetSplit) (*model.AssetSplit, []model.AppliedAssetSplit, error)

	ProcessTdaBuyOrder(ctx context.Context, tx *sql.Tx, input ProcessTdaBuyOrderInput) (*model.Trade, *model.OpenLot, error)
}

type tradeIngestionHandler struct {
}

func NewTradeIngestionService() TradeIngestionService {
	return tradeIngestionHandler{}
}

type ProcessTdaBuyOrderInput struct {
	TdaTransactionID int64
	Symbol           string
	Quantity         decimal.Decimal
	CostBasis        decimal.Decimal
	Date             time.Time
	Description      *string
}

func (h tradeIngestionHandler) ProcessTdaBuyOrder(ctx context.Context, tx *sql.Tx, input ProcessTdaBuyOrderInput) (*model.Trade, *model.OpenLot, error) {
	t := ProcessBuyOrderInput{
		Symbol:      input.Symbol,
		Quantity:    input.Quantity,
		CostBasis:   input.CostBasis,
		Date:        input.Date,
		Description: input.Description,
		Custodian:   model.CustodianType_Tda,
	}

	trade, lots, err := h.ProcessBuyOrder(ctx, tx, t)
	if err != nil {
		return nil, nil, err
	}

	tdaOrder := model.TdaTrade{
		TdaTransactionID: input.TdaTransactionID,
		TradeID:          trade.TradeID,
	}

	err = db.AddTdaTrade(tx, tdaOrder)
	if err != nil && strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "tda_trade_tda_transaction_id_key"`) {
		return nil, nil, hood_errors.ErrDuplicateTrade{
			Custodian:              model.CustodianType_Tda,
			CustodianTransactionID: input.TdaTransactionID,
		}
	} else if err != nil {
		return nil, nil, err
	}

	return trade, lots, nil
}

type ProcessBuyOrderInput struct {
	Symbol      string
	Quantity    decimal.Decimal
	CostBasis   decimal.Decimal
	Date        time.Time
	Description *string
	Custodian   model.CustodianType
}

func (h tradeIngestionHandler) ProcessBuyOrder(ctx context.Context, tx *sql.Tx, in ProcessBuyOrderInput) (*model.Trade, *model.OpenLot, error) {
	t := domain.Trade{
		Symbol:      in.Symbol,
		Quantity:    in.Quantity,
		Price:       in.CostBasis,
		Date:        in.Date,
		Description: in.Description,
		Custodian:   in.Custodian,
		Action:      model.TradeActionType_Buy,
	}

	insertedTrades, err := db.AddTrades(ctx, tx, []domain.Trade{t})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add trades for buy order: %w", err)
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
	insertedLots, err := db.AddOpenLots(ctx, tx, []model.OpenLot{newLot})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, errors.New("no inserted open lots")
	}
	insertedLot := insertedLots[0]

	return &insertedTrade, &insertedLot, nil
}

type ProcessSellOrderInput struct {
	Symbol      string
	Quantity    decimal.Decimal
	CostBasis   decimal.Decimal
	Date        time.Time
	Description *string
	Custodian   model.CustodianType
}

func (h tradeIngestionHandler) ProcessSellOrder(ctx context.Context, tx *sql.Tx, input ProcessSellOrderInput) (*model.Trade, []*model.ClosedLot, error) {
	openLots, err := db.GetOpenLots(ctx, tx, input.Symbol)
	if err != nil {
		return nil, nil, err
	}
	t := domain.Trade{
		Symbol:      input.Symbol,
		Action:      model.TradeActionType_Sell,
		Quantity:    input.Quantity,
		Price:       input.CostBasis,
		Date:        input.Date,
		Description: input.Description,
		Custodian:   input.Custodian,
	}

	insertedTrades, err := db.AddTrades(ctx, tx, []domain.Trade{t})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, nil
	}
	insertedTrade := insertedTrades[0]
	sellOrderResult, err := PreviewSellOrder(t, openLots)
	if err != nil {
		return nil, nil, err
	}
	insertedClosedLots, err := db.AddClosedLots(ctx, tx, sellOrderResult.NewClosedLots)
	if err != nil {
		return nil, nil, err
	}

	return &insertedTrade, insertedClosedLots, nil
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
	NewClosedLots   []domain.ClosedLot
	OpenLots        []domain.OpenLot // current state of open lots
	MutatedOpenLots []domain.OpenLot // lots that were changed
	CashDelta       decimal.Decimal
}

// Selling an asset involves closing currently open lots. In doing this, we may either
// close all open lots for the asset, or close some. The latter requires us to modify
// the existing open lot. Actually, both require us to modify the open lot
//
// This function does the "heavy lifting" to determine which lots should be sold
// without actually selling them. It's only exported because we re-use this logic
// when simulating what a sell order would do
func PreviewSellOrder(t domain.Trade, openLots []domain.OpenLot) (*ProcessSellOrderResult, error) {
	cashDelta := (t.Price.Mul(t.Quantity))
	closedLots := []domain.ClosedLot{}
	mutatedLots := []domain.OpenLot{}
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
		lot.Quantity = lot.Quantity.Sub(quantitySold)
		if lot.Quantity.Equal(decimal.Zero) {
			openLots = openLots[1:]
		}
		modifiedLot := lot.DeepCopy()
		modifiedLot.Date = t.Date
		mutatedLots = append(mutatedLots, *modifiedLot)
		// p.allOpenLots = append(p.allOpenLots, *modifiedLot)

		gains := (t.Price.Sub(lot.CostBasis)).Mul(quantitySold)
		gainsType := model.GainsType_ShortTerm
		daysBetween := t.Date.Sub(lot.GetPurchaseDate())
		if daysBetween.Hours()/24 >= 365 {
			gainsType = model.GainsType_LongTerm
		}
		closedLots = append(closedLots, domain.ClosedLot{
			OpenLot:       &lot,
			SellTrade:     &t,
			Quantity:      quantitySold,
			GainsType:     gainsType,
			RealizedGains: gains,
		})
	}

	return &ProcessSellOrderResult{
		CashDelta:     cashDelta,
		NewClosedLots: closedLots,
	}, nil
}

func (h tradeIngestionHandler) AddAssetSplit(ctx context.Context, tx *sql.Tx, split model.AssetSplit) (*model.AssetSplit, []model.AppliedAssetSplit, error) {
	split.CreatedAt = time.Now().UTC()
	insertedSplits, err := db.AddAssetsSplits(ctx, tx, []*model.AssetSplit{&split})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedSplits) == 0 {
		return nil, nil, nil
	}
	insertedSplit := insertedSplits[0]
	lots, err := db.GetOpenLots(ctx, tx, split.Symbol)
	if err != nil {
		return nil, nil, err
	}

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
