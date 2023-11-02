package metrics

import (
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	"sort"
)

func CalculateDailyPercentChange(prices []model.Price) (map[string]domain.PercentData, error) {
	pricesBySymbol := map[string][]model.Price{}
	// group prices by symbol
	for _, p := range prices {
		if _, ok := pricesBySymbol[p.Symbol]; !ok {
			pricesBySymbol[p.Symbol] = []model.Price{}
		}
		pricesBySymbol[p.Symbol] = append(pricesBySymbol[p.Symbol], p)
	}

	percentChangeBySymbol := map[string]domain.PercentData{}
	for symbol, prices := range pricesBySymbol {
		p, err := PercentChange(prices)
		if err != nil {
			return nil, fmt.Errorf("failed on %s: %w", symbol, err)
		}

		percentChangeBySymbol[symbol] = p
	}

	return percentChangeBySymbol, nil
}

func PercentChange(prices []model.Price) (domain.PercentData, error) {
	if len(prices) < 2 {
		return nil, fmt.Errorf("cannot daily percent change - only %d data points", len(prices))
	}
	sort.SliceStable(prices, func(i, j int) bool {
		return prices[i].Date.Before(prices[j].Date)
	})
	mappedPriceLists := []domain.Percent{}
	for i, p := range prices[1:] {
		prevPrice := prices[i].Price.InexactFloat64()
		currentPrice := p.Price.InexactFloat64()
		percentChange := (currentPrice - prevPrice) / prevPrice
		mappedPriceLists = append(mappedPriceLists, domain.PercentFromFraction(percentChange))
	}
	return mappedPriceLists, nil
}
