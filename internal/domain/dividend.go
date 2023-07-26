package domain

import (
	"hood/internal/db/models/postgres/public/model"
	"time"

	"github.com/shopspring/decimal"
)

type Dividend struct {
	DividendID          *int32
	Symbol              string
	Date                time.Time
	Amount              decimal.Decimal
	Custodian           model.CustodianType
	ReinvestmentTradeID *int32
}

func (d Dividend) GetDate() time.Time { return d.Date }
