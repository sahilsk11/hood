package prices

import (
	"context"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
)

func UpdatePrice(ctx context.Context, priceClient PriceIngestionClient, symbol string) error {
	tx, err := db.GetTx(ctx)
	if err != nil {
		return err
	}

	newPrice, err := priceClient.GetLatestPrice(symbol)
	if err != nil {
		return err
	}
	_, err = db.AddPrices(tx, []model.Price{*newPrice})
	if err != nil {
		return err
	}

	return nil
}

func UpdateCurrentHoldingsPrices(ctx context.Context, priceClient PriceIngestionClient) error {
	tx, err := db.GetTx(ctx)
	if err != nil {
		return err
	}
	holdings, err := db.GetVwHolding(ctx, tx)
	if err != nil {
		return err
	}

	for _, holding := range holdings {
		err = UpdatePrice(ctx, priceClient, *holding.Symbol)
		if err != nil {
			return err
		}
	}
	return nil
}
