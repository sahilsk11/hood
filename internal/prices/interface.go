package prices

import (
	"hood/internal/db/models/postgres/public/model"
	"time"
)

// useless
type PriceIngestionClient interface {
	GetLatestPrice(symbol string) (*model.Price, error)
	GetHistoricalPrices(symbol string, start time.Time) ([]model.Price, error)
}
