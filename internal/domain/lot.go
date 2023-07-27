package domain

import (
	"hood/internal/db/models/postgres/public/model"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OpenLot struct {
	OpenLotID *int32
	LotID     uuid.UUID
	Quantity  decimal.Decimal
	CostBasis decimal.Decimal
	Trade     *Trade
	Date      time.Time
}

type OpenLots []OpenLot

func (lots OpenLots) Ptr() []*OpenLot {
	out := make([]*OpenLot, len(lots))
	for i, o := range lots {
		out[i] = &o
	}
	return out
}

func (o OpenLot) DeepCopy() *OpenLot {
	return &OpenLot{
		OpenLotID: o.OpenLotID,
		LotID:     o.LotID,
		Quantity:  o.Quantity,
		CostBasis: o.CostBasis,
		Trade:     o.Trade.DeepCopy(),
	}
}

func (o OpenLot) GetSymbol() string {
	return o.Trade.Symbol
}

func (o OpenLot) TradeID() *int32 {
	return o.Trade.TradeID
}

func (o OpenLot) GetPurchaseDate() time.Time {
	return o.Trade.Date
}

type ClosedLot struct {
	OpenLot       *OpenLot // not supported yet
	SellTrade     *Trade
	Quantity      decimal.Decimal
	RealizedGains decimal.Decimal
	GainsType     model.GainsType
}

func (c ClosedLot) DeepCopy() ClosedLot {
	openLot := c.OpenLot.DeepCopy()
	sellTrade := c.SellTrade.DeepCopy()
	return ClosedLot{
		OpenLot:       openLot,
		SellTrade:     sellTrade,
		Quantity:      c.Quantity,
		RealizedGains: c.RealizedGains,
		GainsType:     c.GainsType,
	}
}

func (c ClosedLot) Date() time.Time {
	return c.SellTrade.Date
}

func (c ClosedLot) CostBasis() decimal.Decimal {
	// (sell - buy)*quantity = gains
	// sell - buy = gains/quantity
	// purchase price = sell price - realized_gains/closed_lot.quantity
	if c.OpenLot == nil {
		return c.SellTrade.Price.Sub(c.RealizedGains.Div(c.Quantity)).Mul(c.Quantity)
	}
	return c.SellTrade.Price.Sub(c.OpenLot.CostBasis).Mul(c.Quantity)
}
