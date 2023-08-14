package prices

import "hood/internal/db/models/postgres/public/model"

// useless
type PriceIngestionClient interface {
	GetLatestPrice(symbol string) (*model.Price, error)
}
