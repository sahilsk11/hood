package trade_intelligence

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"sort"

	"github.com/shopspring/decimal"
)

type TLHRecomendation struct {
	Symbol               string          `json:"symbol"`
	SellQuantity         decimal.Decimal `json:"sellQuantity"`
	Loss                 decimal.Decimal `json:"loss"`
	BreakevenPriceChange decimal.Decimal `json:"breakevenPriceChange"`
}

func IdentifyTLHOptions(ctx context.Context) ([]TLHRecomendation, error) {
	tlhRecs := []TLHRecomendation{}

	lots, err := db.GetVwOpenLotPosition(ctx)
	if err != nil {
		return nil, err
	}
	lotsByTicker := map[string][]model.VwOpenLotPosition{}
	for _, lot := range lots {
		ticker := *lot.Symbol
		if _, ok := lotsByTicker[ticker]; !ok {
			lotsByTicker[ticker] = []model.VwOpenLotPosition{}
		}
		lotsByTicker[ticker] = append(lotsByTicker[ticker], lot)
	}

	for ticker := range lotsByTicker {
		// not entirely sure, but I think
		// this ordering of ID's sorts the lots in buy order
		sortedLots := lotsByTicker[ticker]
		sort.Slice(sortedLots, func(i, j int) bool {
			return *sortedLots[i].OpenLotID < *sortedLots[j].OpenLotID
		})
		minGain, minGainQuantity := calculateMaxLoss(sortedLots)
		// if we can produce a loss
		if minGain.LessThan(decimal.Zero) {
			price := *sortedLots[0].Price
			breakevenPriceChange := calculateTlhPriceRisk(minGain, minGainQuantity, price)
			tlhRecs = append(tlhRecs, TLHRecomendation{
				Symbol:               ticker,
				SellQuantity:         minGainQuantity,
				Loss:                 minGain.Neg(),
				BreakevenPriceChange: breakevenPriceChange,
			})
		}

	}

	return tlhRecs, nil
}

func calculateMaxLoss(sortedLots []model.VwOpenLotPosition) (decimal.Decimal, decimal.Decimal) {
	minGain := decimal.Zero
	minGainQuantity := decimal.Zero

	currentGain := decimal.Zero
	currentQuantity := decimal.Zero
	for _, lot := range sortedLots {
		currentGain = currentGain.Add(*lot.UnrealizedGains)
		currentQuantity = currentQuantity.Add(*lot.Quantity)
		if currentGain.LessThan(minGain) {
			minGain = currentGain
			minGainQuantity = currentQuantity
		}
	}

	return minGain, minGainQuantity
}

func calculateTlhPriceRisk(minGain, quantity, assetPrice decimal.Decimal) decimal.Decimal {
	ltTaxRate := decimal.NewFromFloat(0.15)

	moneySaved := minGain.Neg().Mul(ltTaxRate)
	// to make that same amount of money,
	// the value of assets sold needs to go
	// up by the same amount AFTER tax
	// 0.85x = moneySaved; x needed to break even
	breakEvenAmount := moneySaved.Div(decimal.NewFromInt(1).Sub(ltTaxRate))
	increasePerUnit := breakEvenAmount.Div(quantity)

	percentChangePerUnit := (increasePerUnit.Div(assetPrice)).Mul(decimal.NewFromInt(100))

	return percentChangePerUnit
}
