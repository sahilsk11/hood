package domain

import (
	"hood/internal/db/models/postgres/public/model"
	"time"

	"github.com/shopspring/decimal"
)

type OpenLot struct {
	OpenLotID    int32
	TradeID      int32
	Symbol       string
	Quantity     decimal.Decimal
	CostBasis    decimal.Decimal
	PurchaseDate time.Time
}

func OpenLotFromVwOpenLotPosition(lot model.VwOpenLotPosition) OpenLot {
	return OpenLot{
		OpenLotID:    *lot.OpenLotID,
		TradeID:      *lot.TradeID,
		Symbol:       *lot.Symbol,
		Quantity:     *lot.Quantity,
		CostBasis:    *lot.CostBasis,
		PurchaseDate: *lot.PurchaseDate,
	}
}
