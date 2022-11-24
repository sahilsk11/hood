package domain

import (
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
