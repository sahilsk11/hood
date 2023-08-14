package prices

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"strings"
	"time"

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

func UpdateHistoric(tx *sql.Tx, priceClient PriceIngestionClient, symbols []string, start time.Time) error {
	for _, s := range symbols {
		savepoint := ""
		prices, err := priceClient.GetHistoricalPrices(s, start)
		if err != nil {
			return db.RollbackWithError(tx, savepoint, err)
		}
		_, err = db.AddPrices(tx, prices)
		if err != nil {
			return db.RollbackWithError(tx, savepoint, err)
		}
		savepoint, err = db.AddSavepoint(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateHistoricMetrics(tx *sql.Tx, djClient DataJockeyClient, symbols []string) error {
	for _, s := range symbols {
		savepoint := ""
		metrics, err := djClient.GetAssetMetrics(s)
		if err != nil {
			return db.RollbackWithError(tx, savepoint, err)
		}
		bytes, err := json.Marshal(metrics)
		if err != nil {
			return db.RollbackWithError(tx, savepoint, err)

		}
		ml := model.DataJockeyAssetMetrics{
			CreatedAt: time.Now().UTC(),
			JSON:      string(bytes),
		}
		err = db.AddDjMetrics(tx, []model.DataJockeyAssetMetrics{ml})
		if err != nil {
			return db.RollbackWithError(tx, savepoint, err)
		}
		savepoint, err = db.AddSavepoint(tx)
		if err != nil {
			return err
		}
	}

	return nil
}
