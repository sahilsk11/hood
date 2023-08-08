package domain

import (
	"hood/internal/db/models/postgres/public/model"
	"time"

	"github.com/shopspring/decimal"
)

type TradeEvent interface {
	GetDate() time.Time
}

type Trade struct {
	TradeID     *int32
	Symbol      string
	Quantity    decimal.Decimal
	Price       decimal.Decimal
	Date        time.Time
	Description *string
	Custodian   model.CustodianType
	Action      model.TradeActionType
}

func (t Trade) DeepCopy() *Trade {
	return &Trade{
		TradeID:     t.TradeID,
		Symbol:      t.Symbol,
		Quantity:    t.Quantity,
		Price:       t.Price,
		Date:        t.Date,
		Description: t.Description,
		Custodian:   t.Custodian,
		Action:      t.Action,
	}
}

func (t Trade) GetDate() time.Time { return t.Date }

func (t Trade) Ptr() *Trade { return &t }

type AssetSplit struct {
	AssetSplitID *int32
	Symbol       string
	Ratio        int32
	Date         time.Time
}

func (t AssetSplit) GetDate() time.Time { return t.Date }

type Transfer struct {
	ActivityID   *int32
	Amount       decimal.Decimal
	ActivityType model.BankActivityType
	Date         time.Time
	Custodian    model.CustodianType
}

func (t Transfer) GetDate() time.Time { return t.Date }

type ProposedTrade struct {
	Symbol        string
	Quantity      decimal.Decimal // negative is valid and implies sell
	ExpectedPrice decimal.Decimal
}

type ProposedTrades []ProposedTrade

func (pts ProposedTrades) ToTrades(date time.Time) []Trade {
	trades := []Trade{}
	for _, pt := range pts {
		quantity := pt.Quantity.Abs()
		action := model.TradeActionType_Buy
		if quantity.LessThan(decimal.Zero) {
			action = model.TradeActionType_Sell
		}
		trades = append(trades, Trade{
			Symbol:   pt.Symbol,
			Quantity: pt.Quantity,
			Price:    pt.ExpectedPrice,
			Date:     date,
			Action:   action,
		})
	}
	return trades
}
