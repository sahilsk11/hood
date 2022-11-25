package price_ingestion

import "hood/internal/db/models/postgres/public/model"

type PriceIngestionClient interface {
	GetLatestPrice(symbol string) (*model.Price, error)
}
