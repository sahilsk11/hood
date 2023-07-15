package portfolio_simulation

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/trade"

	db "hood/internal/db/query"

	"github.com/shopspring/decimal"
)

type SimulateTradeResult struct {
	LongTermGains  decimal.Decimal
	ShortTermGains decimal.Decimal
}

func SimulateTrade(ctx context.Context, t model.Trade) (*SimulateTradeResult, error) {
	tx, err := db.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	openLots, err := db.GetOpenLots(ctx, tx, t.Symbol)
	if err != nil {
		return nil, err
	}
	sellResult, err := trade.PreviewSellOrder(t, openLots)
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
