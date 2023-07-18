package metrics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/trade"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

const PortfolioInception = "2020-06-19"

func CalculateNetReturns(tx *sql.Tx) (decimal.Decimal, error) {
	totalRealizedGains, err := db.GetTotalRealizedGains(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total realized gains: %w", err)
	}
	totalRealizedCostBasis, err := db.GetTotalRealizedCostBasis(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total realized cost basis: %w", err)
	}
	totalUnrealizedGains, err := db.GetTotalUnrealizedGains(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total unrealized gains: %w", err)
	}
	totalUnrealizedCostBasis, err := db.GetTotalUnrealizedCostBasis(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total unrealized cost basis: %w", err)
	}

	totalGains := totalUnrealizedGains.Add(totalRealizedGains)
	totalCostBasis := totalUnrealizedCostBasis.Add(totalRealizedCostBasis)
	if totalCostBasis.Equal(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("received 0 total cost basis: %w", err)
	}
	details := map[string]float64{
		"netRealizedGains":         totalRealizedGains.InexactFloat64(),
		"netUnrealizedGains":       totalUnrealizedGains.InexactFloat64(),
		"closedPositionsCostBasis": totalRealizedCostBasis.InexactFloat64(),
		"openPositionsCostBasis":   totalUnrealizedCostBasis.InexactFloat64(),
		"totalGains":               (totalRealizedGains.Add(totalUnrealizedGains)).InexactFloat64(),
		"totalCostBasis":           (totalRealizedCostBasis.Add(totalUnrealizedCostBasis)).InexactFloat64(),
	}

	b, _ := json.MarshalIndent(details, "", "    ")
	fmt.Println(string(b))

	return totalGains.Div(totalCostBasis), nil
}

func CalculateNetRealizedReturns(tx *sql.Tx) (decimal.Decimal, error) {
	totalRealizedGains, err := db.GetTotalRealizedGains(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total realized gains: %w", err)
	}
	totalRealizedCostBasis, err := db.GetTotalRealizedCostBasis(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total realized cost basis: %w", err)
	}

	if totalRealizedCostBasis.Equal(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("received 0 total cost basis: %w", err)
	}

	return totalRealizedGains.Div(totalRealizedCostBasis), nil
}

func CalculateNetUnrealizedReturns(tx *sql.Tx) (decimal.Decimal, error) {
	totalUnrealizedGains, err := db.GetTotalUnrealizedGains(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total unrealized gains: %w", err)
	}
	totalUnrealizedCostBasis, err := db.GetTotalUnrealizedCostBasis(tx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total unrealized cost basis: %w", err)
	}

	if totalUnrealizedCostBasis.Equal(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("received 0 total cost basis: %w", err)
	}

	return totalUnrealizedGains.Div(totalUnrealizedCostBasis), nil
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

type Portfolio struct {
	OpenLots    map[string][]*domain.OpenLot
	Cash        decimal.Decimal
	NetCashFlow decimal.Decimal
}

func (p Portfolio) deepCopy() Portfolio {
	newP := Portfolio{
		OpenLots:    make(map[string][]*domain.OpenLot),
		Cash:        p.Cash,
		NetCashFlow: p.NetCashFlow,
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
	value = value.Add(p.Cash)

	return value, nil
}

func (p *Portfolio) processTrade(t model.Trade, openLotID *int32) error {
	if t.Action == model.TradeActionType_Buy {
		_, ok := p.OpenLots[t.Symbol]
		if !ok {
			p.OpenLots[t.Symbol] = []*domain.OpenLot{}
		}
		p.OpenLots[t.Symbol] = append(p.OpenLots[t.Symbol], &domain.OpenLot{
			OpenLotID: *openLotID,
			CostBasis: t.CostBasis,
			Quantity:  t.Quantity,
			Trade:     &t,
		})
		*openLotID++
	}
	if t.Action == model.TradeActionType_Sell {
		out, err := trade.PreviewSellOrder(t, p.OpenLots[t.Symbol])
		if err != nil {
			return err
		}
		p.Cash = p.Cash.Add(t.Quantity.Mul(t.CostBasis))

		for _, l := range out.UpdatedOpenLots {
			openLots := p.OpenLots[t.Symbol]
			for i := len(openLots) - 1; i >= 0; i-- {
				if l.OpenLotID == openLots[i].OpenLotID {
					openLots[i].Quantity = l.Quantity

					if l.Quantity.Equal(decimal.Zero) {
						openLots = append(openLots[:i], openLots[i+1:]...)
					}
				}
			}
			p.OpenLots[t.Symbol] = openLots
		}
	}
	return nil
}

func CalculateDailyPortfolioValues(trades []model.Trade, assetSplits []model.AssetSplit, transfers []model.BankActivity, startTime time.Time, endTime time.Time) (map[string]Portfolio, error) {
	p := Portfolio{
		OpenLots: make(map[string][]*domain.OpenLot),
		Cash:     decimal.Zero,
	}
	openLotID := int32(0)
	out := map[string]Portfolio{}

	t := startTime

	for t.Before(endTime) && len(trades) > 0 {
		tomorrow := t.Add(time.Hour * 24)

		// determine relevant models
		relevantTrades := []model.Trade{}
		for len(trades) > 0 && trades[0].Date.Before(tomorrow) {
			relevantTrades = append(relevantTrades, trades[0])
			trades = trades[1:]
		}
		// TODO - edge case of trades and asset splits on the same day
		// we should build a session replayer
		relevantAssetSplits := []model.AssetSplit{}
		for len(assetSplits) > 0 && assetSplits[0].Date.Before(tomorrow) {
			relevantAssetSplits = append(relevantAssetSplits, assetSplits[0])
			assetSplits = assetSplits[1:]
		}
		relevantTransfers := []model.BankActivity{}
		for len(relevantTransfers) > 0 && relevantTransfers[0].Date.Before(tomorrow) {
			relevantTransfers = append(relevantTransfers, relevantTransfers[0])
			relevantTransfers = relevantTransfers[1:]
		}

		// process relevant data
		for _, t := range relevantTransfers {
			p.Cash = p.Cash.Add(t.Amount)
			p.NetCashFlow = p.NetCashFlow.Add(t.Amount)
		}
		for _, split := range relevantAssetSplits {
			ratio := decimal.NewFromInt32(split.Ratio)
			for _, o := range p.OpenLots[split.Symbol] {
				o.CostBasis = o.CostBasis.Div(ratio)
				o.Quantity = o.Quantity.Mul(ratio)
			}
		}
		for _, t := range relevantTrades {
			p.processTrade(t, &openLotID)
		}

		if p.Cash.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("cash below $0 (%f) on %s", p.Cash.InexactFloat64(), t.Format("2006-01-02"))
		}

		out[t.Format("2006-01-02")] = p.deepCopy()
		p.NetCashFlow = decimal.Zero
		t = tomorrow
	}

	return out, nil
}

func TimeWeightedReturns(tx *sql.Tx, dailyPortfolios map[string]Portfolio) (map[string]decimal.Decimal, error) {
	if len(dailyPortfolios) < 2 {
		return nil, fmt.Errorf("at least two daily portfolios required to compute TWR")
	}
	out := map[string]decimal.Decimal{}
	twr := decimal.NewFromInt(1)

	dateKeys := []string{}
	for dateStr := range dailyPortfolios {
		dateKeys = append(dateKeys, dateStr)
	}
	sort.Strings(dateKeys)

	dateKeys = dateKeys[1:]

	for _, dateStr := range dateKeys {
		portfolio := dailyPortfolios[dateStr]
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}
		symbols := []string{}
		for symbol := range portfolio.OpenLots {
			symbols = append(symbols, symbol)
		}

		priceDate := t
		for int(priceDate.Weekday()) == 6 || int(priceDate.Weekday()) == 0 {
			priceDate = priceDate.AddDate(0, 0, -1)
		}

		priceMap, err := getPricesHelper(tx, t)
		if err != nil {
			return nil, err
		}

		yday := t.AddDate(0, 0, -1)
		ydayPriceMap, err := getPricesHelper(tx, yday)
		if err != nil {
			return nil, err
		}

		ydayPortfolio, ok := dailyPortfolios[yday.Format("2006-01-02")]
		if !ok {
			return nil, fmt.Errorf("could not find yesterdays portfolio on %s", yday.Format("2006-01-02"))
		}

		start, err := ydayPortfolio.netValue(ydayPriceMap)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate net value on %s: %w", yday.Format("2006-01-02"), err)
		}
		end, err := portfolio.netValue(priceMap)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate net value on %s: %w", yday.Format(dateStr), err)
		}

		// https://www.investopedia.com/terms/t/time-weightedror.asp

		hp := (end.Sub(start).Sub(portfolio.NetCashFlow)).Div(start.Add(portfolio.NetCashFlow))
		fmt.Println(end.Sub(start).Sub(portfolio.NetCashFlow), start.Add(portfolio.NetCashFlow))
		fmt.Println("hp", hp, "start", start, "end", end, "flow", portfolio.NetCashFlow, "date", dateStr)
		fmt.Println(hp)

		newOp := decimal.NewFromInt(1).Add(hp)
		out[dateStr] = twr.Mul(newOp).Sub(decimal.NewFromInt(1))
		twr = twr.Mul(newOp)
	}

	return out, nil
}

func getPricesHelper(tx *sql.Tx, date time.Time) (map[string]decimal.Decimal, error) {
	priceMap, err := db.GetPricesOnDate(tx, date, []string{})
	if err != nil {
		e := err
		tries := 3
		for tries > 0 && e != nil {
			date = date.AddDate(0, 0, -1)
			priceMap, e = db.GetPricesOnDate(tx, date, []string{})
			tries -= 1
		}
		if e != nil {
			return nil, err
		}
	}

	return priceMap, err
}
