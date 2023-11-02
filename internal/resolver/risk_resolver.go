package resolver

import (
	"database/sql"
	"fmt"
	api "hood/api-types"
	"hood/internal/service"
)

type Resolver struct {
	Db *sql.DB
}

func (r Resolver) PortfolioCorrelation(req api.PortfolioCorrelationRequest) (*api.PortfolioCorrelationResponse, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return nil, err
	}

	if len(req.Symbols) < 2 {
		return nil, fmt.Errorf("must provide more than two symbols")
	}

	corrs, err := service.PortfolioCorrelation(tx, req.Symbols)
	if err != nil {
		return nil, err
	}

	outputCorrs := []api.Correlation{}
	for _, c := range corrs {
		outputCorrs = append(outputCorrs, api.Correlation{
			AssetOne:    c.AssetOne,
			AssetTwo:    c.AssetTwo,
			Correlation: c.Correlation,
		})
	}

	err = tx.Rollback()
	if err != nil {
		return nil, err
	}

	return &api.PortfolioCorrelationResponse{
		Correlations: outputCorrs,
	}, nil
}
