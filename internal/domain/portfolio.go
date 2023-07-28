package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Portfolio struct {
	OpenLots   map[string][]*OpenLot
	ClosedLots map[string][]ClosedLot
	Cash       decimal.Decimal
	LastAction time.Time
}

func (p Portfolio) DeepCopy() Portfolio {
	out := Portfolio{
		OpenLots:   map[string][]*OpenLot{},
		ClosedLots: map[string][]ClosedLot{},
		Cash:       p.Cash,
		LastAction: p.LastAction,
	}
	for k, v := range p.OpenLots {
		t := []*OpenLot{}
		for _, x := range v {
			t = append(t, x.DeepCopy())
		}
		out.OpenLots[k] = t
	}
	for k, v := range p.ClosedLots {
		t := []ClosedLot{}
		for _, x := range v {
			t = append(t, x.DeepCopy())
		}
		out.ClosedLots[k] = t
	}

	return out
}

func (p Portfolio) GetOpenLotSymbols() []string {
	symbolsMap := map[string]struct{}{}
	for s := range p.OpenLots {
		symbolsMap[s] = struct{}{}
	}
	symbols := []string{}
	for s := range symbolsMap {
		symbols = append(symbols, s)
	}
	return symbols
}

type ValueType string

const (
	PortfolioValueType_Twr    ValueType = "TWR"
	PortfolioValueType_Dollar ValueType = "DOLLAR"
)

// Anything of type "Daily_XX" is expected to contain
// values for a particular field for every day over
// the range INCLUDING weekends

type DailyValue struct {
	Date  time.Time
	Value decimal.Decimal
}

type DailyValues struct {
	values []DailyValue
	start  time.Time
	end    time.Time
}

func NewDailyValues(map[string]decimal.Decimal) (*DailyValues, error) {

}
