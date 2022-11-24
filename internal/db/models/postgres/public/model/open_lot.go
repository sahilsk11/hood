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

type OpenLot struct {
	OpenLotID  int32 `sql:"primary_key"`
	CostBasis  decimal.Decimal
	Quantity   decimal.Decimal
	TradeID    int32
	DeletedAt  *time.Time
	CreatedAt  time.Time
	ModifiedAt time.Time
}
