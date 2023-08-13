package prices

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"strings"
	"time"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
	"github.com/shopspring/decimal"
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

func determineColumnOrdering(headerRow, requiredHeaders []string) (map[string]int, error) {
	headerIndex := map[string]int{}
	for i, h := range headerRow {
		for _, rh := range requiredHeaders {
			h = strings.TrimPrefix(h, "\xef\xbb\xbf")
			if strings.EqualFold(h, rh) {
				headerIndex[rh] = i
			}
		}
	}

	for _, rh := range requiredHeaders {
		if _, ok := headerIndex[rh]; !ok {
			return nil, fmt.Errorf("csv missing required header %s", rh)
		}
	}

	return headerIndex, nil
}

func UpdateFromCsv(tx *sql.Tx, records [][]string) error {
	requiredHeaders := []string{"symbol", "price", "date"}
	headerIndex, err := determineColumnOrdering(records[0], requiredHeaders)
	if err != nil {
		return err
	}

	prices := []model.Price{}
	for _, row := range records[1:] {
		price, err := decimal.NewFromString(row[headerIndex["price"]])
		if err != nil {
			return err
		}
		date, err := time.Parse("2006-01-02", row[headerIndex["date"]])
		if err != nil {
			return err
		}

		p := model.Price{
			Symbol: row[headerIndex["symbol"]],
			Date:   date,
			Price:  price,
		}

		prices = append(prices, p)
	}
	_, err = db.AddPrices(tx, prices)
	if err != nil {
		return err
	}

	return nil
}

// use db bc if something fails
// i want to commit everything
func UpdateFromYahoo(dbThing *sql.DB) error {
	symbols := GetUniverseAssets()
	for _, s := range symbols {
		q, err := equity.Get(s)
		if err != nil {
			return fmt.Errorf("failed to get equity data from yahoo api: %w", err)
		}
		err = db.AddAssetMetrics(dbThing, []model.AssetMetric{
			dbAssetMetricFromYahoo(q),
		})
		if err != nil {
			return err
		}
		fmt.Printf("added %s\n", s)
	}

	return nil
}

func dbAssetMetricFromYahoo(e *finance.Equity) model.AssetMetric {
	return model.AssetMetric{
		Symbol:                                  e.Symbol,
		FullName:                                e.LongName,
		Price:                                   decimal.NewFromFloat(e.RegularMarketPrice),
		PriceUpdatedAt:                          time.Unix(int64(e.RegularMarketTime), 0),
		EarningsPerShareAnnualTrailing:          decimal.NewFromFloat(e.EpsTrailingTwelveMonths),
		EarningsPerShareAnnualTrailingUpdatedAt: time.Unix(int64(e.EarningsTimestamp), 0),
		DividendYieldAnnualTrailing:             decimal.NewFromFloat(e.TrailingAnnualDividendRate),
		DividendYieldAnnualTrailingUpdatedAt:    time.Unix(int64(e.DividendDate), 0),
		PeRatioTrailing:                         decimal.NewFromFloat(e.TrailingPE),
		BookValue:                               decimal.NewFromFloat(e.BookValue),
		PriceToBookRatio:                        decimal.NewFromFloat(e.PriceToBook),
		SharesOutstanding:                       decimal.NewFromInt(int64(e.SharesOutstanding)),
		MarketCap:                               decimal.NewFromInt(e.MarketCap),
	}
}
