package metrics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/trade"
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

type Portfolio struct {
	OpenLots   map[string][]*domain.OpenLot
	ClosedLots map[string][]*domain.ClosedLot
}

func (p Portfolio) deepCopy() Portfolio {
	newP := Portfolio{
		OpenLots:   make(map[string][]*domain.OpenLot),
		ClosedLots: make(map[string][]*domain.ClosedLot),
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
	for k, v := range p.ClosedLots {
		for _, o := range v {
			if _, ok := newP.ClosedLots[k]; !ok {
				newP.ClosedLots[k] = []*domain.ClosedLot{}
			}
			newP.ClosedLots[k] = append(newP.ClosedLots[k], &domain.ClosedLot{
				// BuyTrade:      o.BuyTrade,
				SellTrade:     o.SellTrade,
				Quantity:      o.Quantity,
				RealizedGains: o.RealizedGains,
				GainsType:     o.GainsType,
			})
		}
	}
	return newP
}

func (p Portfolio) CalculateReturns(priceMap map[string]decimal.Decimal) (decimal.Decimal, error) {
	openLotsCostBasis := decimal.Zero
	openLotsGains := decimal.Zero
	closedLotsCostBasis := decimal.Zero
	closedLotsGains := decimal.Zero

	if len(p.ClosedLots) == 0 && len(p.OpenLots) == 0 {
		return decimal.Zero, fmt.Errorf("cannot calculate returns on portfolio with no trades")
	}

	for symbol, lots := range p.OpenLots {
		gains := decimal.Zero
		costBasis := decimal.Zero
		price, ok := priceMap[symbol]
		if !ok {
			return decimal.Zero, fmt.Errorf("missing pricing for %s", symbol)
		}
		for _, lot := range lots {
			costBasis = costBasis.Add((lot.CostBasis.Mul(lot.Quantity)))
			gains = gains.Add(price.Sub(lot.CostBasis).Mul(lot.Quantity))
		}
		openLotsCostBasis = openLotsCostBasis.Add(costBasis)
		openLotsGains = openLotsGains.Add(gains)
	}
	for _, v := range p.ClosedLots {
		for _, lot := range v {
			closedLotsCostBasis = closedLotsCostBasis.Add(lot.CostBasis())
			closedLotsGains = closedLotsGains.Add(lot.RealizedGains)
		}
	}

	totalCostBasis := openLotsCostBasis.Add(closedLotsCostBasis)
	totalGains := openLotsGains.Add(closedLotsGains)

	fmt.Println(openLotsCostBasis.String())
	details := map[string]float64{
		"netRealizedGains":         totalGains.InexactFloat64(),
		"netUnrealizedGains":       openLotsGains.InexactFloat64(),
		"closedPositionsCostBasis": closedLotsCostBasis.InexactFloat64(),
		"openPositionsCostBasis":   openLotsCostBasis.InexactFloat64(),
		"totalGains":               totalGains.InexactFloat64(),
		"totalCostBasis":           totalCostBasis.InexactFloat64(),
	}
	bytes, err := json.Marshal(details)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))

	if totalCostBasis.Equal(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("zero total cost basis")
	}
	fmt.Println(totalGains, totalCostBasis)
	return totalGains.Div(totalCostBasis), nil
}

func DailyReturns(trades []model.Trade, assetSplits []model.AssetSplit, startTime time.Time, endTime time.Time) (map[string]Portfolio, error) {
	p := Portfolio{
		OpenLots:   make(map[string][]*domain.OpenLot),
		ClosedLots: make(map[string][]*domain.ClosedLot),
	}
	openLotID := int32(0)
	out := map[string]Portfolio{}

	t := startTime
	for t.Before(endTime) && len(trades) > 0 {
		tomorrow := t.Add(time.Hour * 24)
		if int(t.Weekday()) != 6 && int(t.Weekday()) != 0 {
			relevantTrades := []model.Trade{}
			for len(trades) > 0 && trades[0].Date.Before(tomorrow) {
				relevantTrades = append(relevantTrades, trades[0])
				trades = trades[1:]
			}
			// TODO - edge case of trades and asset splits on the same day
			relevantAssetSplits := []model.AssetSplit{}
			for len(assetSplits) > 0 && assetSplits[0].Date.Before(tomorrow) {
				relevantAssetSplits = append(relevantAssetSplits, assetSplits[0])
				assetSplits = assetSplits[1:]
			}

			for _, split := range relevantAssetSplits {
				ratio := decimal.NewFromInt32(split.Ratio)
				for _, o := range p.OpenLots[split.Symbol] {
					o.CostBasis = o.CostBasis.Div(ratio)
					o.Quantity = o.Quantity.Mul(ratio)
				}
			}

			for _, t := range relevantTrades {
				if t.Action == model.TradeActionType_Buy {
					_, ok := p.OpenLots[t.Symbol]
					if !ok {
						p.OpenLots[t.Symbol] = []*domain.OpenLot{}
					}
					p.OpenLots[t.Symbol] = append(p.OpenLots[t.Symbol], &domain.OpenLot{
						OpenLotID: openLotID,
						CostBasis: t.CostBasis,
						Quantity:  t.Quantity,
						Trade:     t,
					})
					openLotID++
				}
				if t.Action == model.TradeActionType_Sell {
					out, err := trade.PreviewSellOrder(t, p.OpenLots[t.Symbol])
					if err != nil {
						return nil, err
					}
					p.ClosedLots[t.Symbol] = append(p.ClosedLots[t.Symbol], out.NewDomainClosedLots...)

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
			}
			out[t.Format("2006-01-02")] = p.deepCopy()
		}
		t = tomorrow
	}

	return out, nil
}
