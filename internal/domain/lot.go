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
	Trade        *model.Trade
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

type ClosedLot struct {
	// BuyTrade      model.Trade // not supported yet
	SellTrade     *model.Trade
	Quantity      decimal.Decimal
	RealizedGains decimal.Decimal
	GainsType     model.GainsType
}

func (c ClosedLot) Date() time.Time {
	return c.SellTrade.Date
}

func (c ClosedLot) CostBasis() decimal.Decimal {
	// (sell - buy)*quantity = gains
	// sell - buy = gains/quantity
	// purchase price = sell price - realized_gains/closed_lot.quantity
	return c.SellTrade.CostBasis.Sub(c.RealizedGains.Div(c.Quantity)).Mul(c.Quantity)
}
