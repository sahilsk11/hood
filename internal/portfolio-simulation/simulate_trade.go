package portfolio_simulation

import (
	"context"
	data_ingestion "hood/internal/data-ingestion"
	"hood/internal/db/models/postgres/public/model"

	"github.com/shopspring/decimal"
)

type SimulateTradeResult struct {
	LongTermGains  decimal.Decimal
	ShortTermGains decimal.Decimal
}

func SimulateTrade(ctx context.Context, trade model.Trade) (*SimulateTradeResult, error) {
	openLots, err := GetOpenLotsFromDb(ctx, trade.Symbol)
	if err != nil {
		return nil, err
	}
	sellResult, err := data_ingestion.ProcessSellOrder(trade, openLots)
	if err != nil {
		return nil, err
	}

	var (
		shortTermGains = decimal.Zero
		longTermGains  = decimal.Zero
	)
	for _, lot := range sellResult.NewClosedLots {
		if lot.GainsType == model.GainsType_ShortTerm {
			shortTermGains = shortTermGains.Add(lot.RealizedGains)
		} else {
			longTermGains = longTermGains.Add(lot.RealizedGains)
		}
	}

	return &SimulateTradeResult{
		ShortTermGains: shortTermGains,
		LongTermGains:  longTermGains,
	}, nil
}
