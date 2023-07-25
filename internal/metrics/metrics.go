package metrics

import (
	"database/sql"
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/util"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

const PortfolioInception = "2020-06-19"

// goal is to figure out portfolio value
// at every day from [1:] in days (2 days min)
// day interval calculated as start of day being
// t-1 close (when pricing is avail) to t close
// which should ensure all trades are in
// this is also a proper "trading day"
// since market open will be at prev day close
// some algo should calculate total value at a given day
// call that in loop, generate arr/map of days and value
// then simple func computes on that DS using equation
// and produces arr/map of returns on given day

type Portfolio struct {
	OpenLots        map[string][]*domain.OpenLot
	Transfer        decimal.Decimal
	NetTransferFlow decimal.Decimal
}

func (p Portfolio) deepCopy() Portfolio {
	newP := Portfolio{
		OpenLots:        make(map[string][]*domain.OpenLot),
		Transfer:        p.Transfer,
		NetTransferFlow: p.NetTransferFlow,
	}
	for k, v := range p.OpenLots {
		for _, o := range v {
			if _, ok := newP.OpenLots[k]; !ok {
				newP.OpenLots[k] = []*domain.OpenLot{}
			}
			newP.OpenLots[k] = append(newP.OpenLots[k], &domain.OpenLot{
				OpenLotID: o.OpenLotID,
				Quantity:  o.Quantity,
				CostBasis: o.CostBasis,
				Trade:     o.Trade,
			})
		}
	}
	return newP
}

func (p Portfolio) symbols() []string {
	symbols := []string{}
	for symbol := range p.OpenLots {
		symbols = append(symbols, symbol)
	}
	return symbols
}

func (p Portfolio) netValue(priceMap map[string]decimal.Decimal) (decimal.Decimal, error) {
	value := decimal.Zero
	for symbol, lots := range p.OpenLots {
		price, ok := priceMap[symbol]
		if !ok && symbol != "AMAG" && symbol != "ETH" && symbol != "BTC" && symbol != "DOGE" {
			return decimal.Zero, fmt.Errorf("missing pricing for %s", symbol)
		}
		for _, lot := range lots {
			if symbol == "AMAG" || symbol == "ETH" || symbol == "BTC" || symbol == "DOGE" {
				price = lot.CostBasis
			}
			value = value.Add(price.Mul(lot.Quantity))
		}
	}
	value = value.Add(p.Transfer)

	return value, nil
}

func TimeWeightedReturns(dailyPortfolioValues map[string]decimal.Decimal, transfers map[string]decimal.Decimal) (map[string]decimal.Decimal, error) {
	if len(dailyPortfolioValues) < 2 {
		return nil, fmt.Errorf("at least two daily portfolios required to compute TWR")
	}
	out := map[string]decimal.Decimal{}
	twr := decimal.NewFromInt(1)

	dateKeys := []string{}
	for dateStr := range dailyPortfolioValues {
		dateKeys = append(dateKeys, dateStr)
	}
	sort.Strings(dateKeys)

	dateKeys = dateKeys[1:]
	for _, dateStr := range dateKeys {
		today, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}
		yday := today.AddDate(0, 0, -1)
		end, ok := dailyPortfolioValues[dateStr]
		if !ok {
			return nil, fmt.Errorf("failed to calculate net value - no calculated portfolio value on %s", dateStr)
		}
		start, ok := dailyPortfolioValues[yday.Format("2006-01-02")]
		if !ok {
			return nil, fmt.Errorf("failed to calculate net value - no calculated portfolio value on %s", yday.Format("2006-01-02"))
		}
		netTransfers, _ := transfers[dateStr]

		newOp := hp(start, end, netTransfers)

		out[dateStr] = twr.Mul(newOp).Sub(decimal.NewFromInt(1))
		twr = twr.Mul(newOp)
	}

	return out, nil
}

// https://www.investopedia.com/terms/t/time-weightedror.asp
func hp(start, end, cashFlow decimal.Decimal) decimal.Decimal {
	numerator := end
	denominator := start.Add(cashFlow)
	util.Pprint(map[string]decimal.Decimal{
		"hp":          numerator.Div(denominator),
		"numerator":   numerator,
		"denominator": denominator,
		"start":       start,
		"end":         end,
		"cashFlows":   cashFlow,
	})

	return numerator.Div(denominator)
}

func getPricesHelper(tx *sql.Tx, date time.Time, symbols []string) (map[string]decimal.Decimal, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("getPricesHelper requires at least one symbol")
	}
	priceMap, err := db.GetPricesOnDate(tx, date, symbols)
	if err != nil {
		e := err
		tries := 3
		for tries > 0 && e != nil {
			date = date.AddDate(0, 0, -1)
			priceMap, e = db.GetPricesOnDate(tx, date, symbols)
			tries -= 1
		}
		if e != nil {
			return nil, fmt.Errorf("failed to get prices: %w", err)
		}
	}

	return priceMap, nil
}
