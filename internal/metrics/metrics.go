package metrics

import (
	"database/sql"
	"fmt"
	db "hood/internal/db/query"
	. "hood/internal/domain"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

const PortfolioInception = "2020-06-19"
const layout = "2006-01-02"

func netValue(p Portfolio, priceMap map[string]decimal.Decimal) (decimal.Decimal, error) {
	// value = (lot quantity * price) + cash
	value := p.Cash
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

	return value, nil
}

// determine what the value of the portfolio is on a given day
func calculatePortfolioValue(tx *sql.Tx, p Portfolio, date time.Time) (decimal.Decimal, error) {
	if len(p.GetOpenLotSymbols()) == 0 {
		return p.Cash, nil
	}
	// get prices up to 3 days back
	priceMap, err := getPricesHelper(tx, date, p.GetOpenLotSymbols())
	if err != nil {
		return decimal.Zero, err
	}
	return netValue(p, priceMap)
}

func DailyPortfolioValues(
	tx *sql.Tx,
	dailyPortfolios map[string]Portfolio,
	start *time.Time,
	end *time.Time,
) (map[string]decimal.Decimal, error) {
	if len(dailyPortfolios) == 0 {
		return nil, fmt.Errorf("no portfolios given")
	}
	out := map[string]decimal.Decimal{}
	dateKeys := []string{}

	for dateStr := range dailyPortfolios {
		dateKeys = append(dateKeys, dateStr)
	}
	sort.Strings(dateKeys)
	minPortfolioDate, err := time.Parse(layout, dateKeys[0])
	if err != nil {
		return nil, err
	}
	maxPortfolioDate, err := time.Parse(layout, dateKeys[len(dateKeys)-1])
	if err != nil {
		return nil, err
	}
	if start == nil {
		start = &minPortfolioDate
	}
	if end == nil {
		end = &maxPortfolioDate
	}

	if start.Before(minPortfolioDate) {
		return nil, fmt.Errorf("cannot start calculations prior to date of first portfolio value - %s vs %s", start.Format(layout), minPortfolioDate.Format(layout))
	}

	if end.Before(minPortfolioDate) {
		return nil, fmt.Errorf("inputted end date %s is before first portfolio date %s", end.Format(layout), start.Format(layout))
	}

	// increment portfolio date until we reach
	// start date
	currentTime := minPortfolioDate
	portfolio := dailyPortfolios[dateKeys[0]]
	for currentTime.Before(*start) {
		// if there's a newer portfolio, use it
		if p, ok := dailyPortfolios[currentTime.Format(layout)]; ok {
			portfolio = p
		}
		currentTime = currentTime.AddDate(0, 0, 1)
	}

	for currentTime.Before(*end) || currentTime.Equal(*end) {
		dateStr := currentTime.Format(layout)
		if p, ok := dailyPortfolios[dateStr]; ok {
			portfolio = p
		}

		priceMap, err := getPricesHelper(tx, currentTime, portfolio.GetOpenLotSymbols())
		if err != nil {
			return nil, err
		}
		value, err := netValue(portfolio, priceMap)
		if err != nil {
			return nil, err
		}

		out[dateStr] = value
		currentTime = currentTime.AddDate(0, 0, 1)
	}

	return out, nil
}

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

func TimeWeightedReturns(
	dailyPortfolioValues map[string]decimal.Decimal,
	transfers map[string]decimal.Decimal,
) (map[string]decimal.Decimal, error) {
	if len(dailyPortfolioValues) < 2 {
		return nil, fmt.Errorf("at least two daily portfolios required to compute TWR")
	}
	dateKeys := []string{}
	for dateStr := range dailyPortfolioValues {
		dateKeys = append(dateKeys, dateStr)
	}
	sort.Strings(dateKeys)

	out := map[string]decimal.Decimal{}
	twr := decimal.NewFromInt(1)

	for i := 1; i < len(dateKeys); i++ {
		prevValue := dailyPortfolioValues[dateKeys[i-1]]
		currentValue := dailyPortfolioValues[dateKeys[i]]
		netTransfers, _ := transfers[dateKeys[i]]

		// https://www.investopedia.com/terms/t/time-weightedror.asp
		x := prevValue.Add(netTransfers)
		twr = twr.Mul(
			((currentValue.Sub(x)).Div(x)).Add(decimal.NewFromInt(1)),
		)
		out[dateKeys[i]] = twr.Sub(decimal.NewFromInt(1))
	}

	return out, nil
}

func getPricesHelper(tx *sql.Tx, date time.Time, symbols []string) (map[string]decimal.Decimal, error) {
	if len(symbols) == 0 {
		return map[string]decimal.Decimal{}, nil
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
