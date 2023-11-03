package resolver

import (
	"database/sql"
	"fmt"
	api "hood/api-types"
	"hood/internal/service"
	"sort"
)

type Resolver struct {
	Db *sql.DB
}

func (r Resolver) PortfolioCorrelation(req api.CorrelationMatrixRequest) (*api.CorrelationMatrixResponse, error) {
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

	sort.Slice(outputCorrs, func(i, j int) bool {
		return outputCorrs[i].Correlation > outputCorrs[j].Correlation
	})

	err = tx.Rollback()
	if err != nil {
		return nil, err
	}

	return &api.CorrelationMatrixResponse{
		Correlations: outputCorrs,
	}, nil
}

func (r Resolver) CorrelationAllocation(req api.CorrelationAllocationRequest) (*api.CorrelationAllocationResponse, error) {
	// tx, err := r.Db.Begin()
	// if err != nil {
	// 	return nil, err
	// }
	return nil, nil
}
