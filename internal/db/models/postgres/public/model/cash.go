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

type Cash struct {
	CashID    int32 `sql:"primary_key"`
	Amount    decimal.Decimal
	Custodian CustodianType
	CreatedAt time.Time
	Date      time.Time
}
