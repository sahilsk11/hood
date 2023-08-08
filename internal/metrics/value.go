package metrics

import (
	"database/sql"
	"fmt"
	db "hood/internal/db/query"
	. "hood/internal/domain"
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
// need seperate date param for pricing on date
func CalculatePortfolioValue(tx *sql.Tx, p Portfolio, date time.Time) (decimal.Decimal, error) {
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

// determine what the value of the portfolio is on a given day
func CalculateMetricsPortfolioValue(tx *sql.Tx, mp MetricsPortfolio, date time.Time) (decimal.Decimal, error) {
	if len(mp.Symbols()) == 0 {
		return mp.Cash, nil
	}
	// get prices up to 3 days back
	priceMap, err := getPricesHelper(tx, date, mp.Symbols())
	if err != nil {
		return decimal.Zero, err
	}
	out := mp.Cash
	for symbol, p := range mp.Positions {
		out = out.Add(p.Quantity.Mul(priceMap[symbol]))
	}
	return out, nil
}

// over the given date range, determine
// what the value of a portfolio is on every
// day within the range
func DailyPortfolioValues(
	tx *sql.Tx,
	hp HistoricPortfolio,
	start *time.Time,
	end *time.Time,
) (map[string]decimal.Decimal, error) {
	if len(hp.GetPortfolios()) == 0 {
		return nil, fmt.Errorf("no portfolios given")
	}
	out := map[string]decimal.Decimal{}

	minPortfolioDate := hp.GetPortfolios()[0].LastAction
	maxPortfolioDate := hp.Latest().LastAction
	if start == nil {
		start = &minPortfolioDate
	} else {
		fmt.Println("not configured")
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

	nextHpIndex := 1
	currentTime := *start
	currentPortfolio := hp.GetPortfolios()[0]

	for currentTime.Unix() <= end.Unix() {
		if nextHpIndex < len(hp.GetPortfolios()) && (hp.GetPortfolios()[nextHpIndex].LastAction.Unix() >= currentTime.Unix()) {
			currentPortfolio = hp.GetPortfolios()[nextHpIndex]
			nextHpIndex++
		}
		dateStr := currentTime.Format(layout)

		value, err := CalculatePortfolioValue(
			tx,
			currentPortfolio,
			currentTime,
		)
		if err != nil {
			return nil, err
		}

		out[dateStr] = value
		currentTime = currentTime.AddDate(0, 0, 1)
	}

	// // increment portfolio date until we reach
	// // start date
	// i := 0
	// for i+1 < len(portfolios.GetPortfolios()) && portfolios.GetPortfolios()[i+1].LastAction.Before(*start) {
	// 	i++
	// }

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
