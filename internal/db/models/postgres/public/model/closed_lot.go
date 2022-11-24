//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type ClosedLot struct {
	ClosedLotID   int32 `sql:"primary_key"`
	BuyTradeID    int32
	SellTradeID   int32
	Quantity      decimal.Decimal
	RealizedGains decimal.Decimal
	GainsType     GainsType
	CreatedAt     time.Time
	ModifiedAt    time.Time
}
