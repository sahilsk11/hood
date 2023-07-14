package trade_ingestion

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
	"math"
	"sort"
	"strings"
	"time"

	"hood/internal/util"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// consider changing these names from "Add" to something better

type TradeIngestionService interface {
	ProcessBuyOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade) (*model.Trade, *model.OpenLot, error)
	ProcessSellOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade) (*model.Trade, []*model.ClosedLot, error)
	AddAssetSplit(ctx context.Context, tx *sql.Tx, split model.AssetSplit) (*model.AssetSplit, []model.AppliedAssetSplit, error)

	ProcessTdaBuyOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade, tdaTxId int64) (*model.Trade, *model.OpenLot, error)
}

type tradeIngestionHandler struct {
}

func NewTradeIngestionService(ctx context.Context, tx *sql.Tx) TradeIngestionService {
	return tradeIngestionHandler{}
}

func (h tradeIngestionHandler) ProcessTdaBuyOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade, tdaTxId int64) (*model.Trade, *model.OpenLot, error) {
	if newTrade.Custodian != model.CustodianType_Tda {
		return nil, nil, fmt.Errorf("cannot process tda buy order with custodian %s", newTrade.Custodian.String())
	}

	savepointName := "x" + strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := tx.Exec("SAVEPOINT " + savepointName + ";")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create savepoint at ProcessTdaBuyOrder: %w", err)
	}

	trade, lots, err := h.ProcessBuyOrder(ctx, tx, newTrade)
	if err != nil {
		return nil, nil, err
	}

	tdaOrder := model.TdaTrade{
		TdaTransactionID: tdaTxId,
		TradeID:          &trade.TradeID,
	}

	_, err = table.TdaTrade.INSERT(table.TdaTrade.MutableColumns).MODEL(tdaOrder).Exec(tx)
	if err != nil {
		_, rollbackErr := tx.Exec("ROLLBACK TO SAVEPOINT " + savepointName)
		if rollbackErr != nil {
			return nil, nil, fmt.Errorf("failed to rollback buy order: %w", rollbackErr)
		}
		// don't hate me
		if err.Error() == `pq: duplicate key value violates unique constraint "tda_trade_tda_transaction_id_key"` {
			return nil, nil, hood_errors.ErrDuplicateTrade{
				Custodian:              model.CustodianType_Tda,
				CustodianTransactionID: tdaTxId,
			}
		}
		return nil, nil, err
	}

	return trade, lots, nil
}

func (h tradeIngestionHandler) ProcessBuyOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade) (*model.Trade, *model.OpenLot, error) {
	newTrade.CreatedAt = time.Now().UTC()
	newTrade.ModifiedAt = time.Now().UTC()
	insertedTrades, err := db.AddTrades(ctx, tx, []*model.Trade{&newTrade})
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
	insertedLots, err := db.AddOpenLots(ctx, tx, []*model.OpenLot{&newLot})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, errors.New("no inserted open lots")
	}
	insertedLot := insertedLots[0]

	return &insertedTrade, &insertedLot, nil
}

func (h tradeIngestionHandler) ProcessSellOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade) (*model.Trade, []*model.ClosedLot, error) {
	openLots, err := db.GetOpenLots(ctx, tx, newTrade.Symbol)
	if err != nil {
		return nil, nil, err
	}
	newTrade.CreatedAt = time.Now().UTC()
	newTrade.ModifiedAt = time.Now().UTC()
	insertedTrades, err := db.AddTrades(ctx, tx, []*model.Trade{&newTrade})
	if err != nil {
		return nil, nil, err
	}
	if len(insertedTrades) == 0 {
		return nil, nil, nil
	}
	insertedTrade := insertedTrades[0]
	sellOrderResult, err := PreviewSellOrder(insertedTrade, openLots)
	if err != nil {
		return nil, nil, err
	}
	insertedClosedLots, err := db.AddClosedLots(ctx, tx, sellOrderResult.NewClosedLots)
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
		_, err = db.UpdateOpenLotInDb(ctx, tx, dbOpenLot, columnlist)
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

// Selling an asset involves closing currently open lots. In doing this, we may either
// close all open lots for the asset, or close some. The latter requires us to modify
// the existing open lot. Actually, both require us to modify the open lot
//
// This function does the "heavy lifting" to determine which lots should be sold
// without actually selling them. It's only exported because we re-use this logic
// when simulating what a sell order would do
func PreviewSellOrder(t model.Trade, openLots []*domain.OpenLot) (*ProcessSellOrderResult, error) {
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
			OpenLotID: lot.OpenLotID,
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
