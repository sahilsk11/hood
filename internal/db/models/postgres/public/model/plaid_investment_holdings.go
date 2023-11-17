//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"github.com/google/uuid"
	"time"

	"github.com/shopspring/decimal"
)

type PlaidInvestmentHoldings struct {
	PlaidInvestmentsHoldingsID uuid.UUID `sql:"primary_key"`
	Ticker                     string
	TradingAccountID           uuid.UUID
	TotalCostBasis             decimal.Decimal
	Quantity                   decimal.Decimal
	CreatedAt                  time.Time
}