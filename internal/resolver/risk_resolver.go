package resolver

import (
	"database/sql"
	"fmt"
	api "hood/api-types"
	"hood/internal/domain"
	"hood/internal/service"
	"sort"

	"github.com/shopspring/decimal"
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

func (r Resolver) CorrelatedAssetGroups(req api.CorrelatedAssetGroupsRequest) (*api.CorrelatedAssetGroupsResponse, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return nil, err
	}

	if len(req.Holdings) < 2 {
		return nil, fmt.Errorf("cannot calculated correlated asset groups with < 2 positions")
	}

	portfolio := domain.MetricsPortfolio{
		Positions: make(map[string]*domain.Position),
	}
	for _, holding := range req.Holdings {
		portfolio.Positions[holding.Symbol] = &domain.Position{
			Symbol:   holding.Symbol,
			Quantity: decimal.NewFromFloat(holding.Quantity),
		}
	}

	assetGroups, err := service.CalculateCorrelatedAssetGroups(tx, portfolio)
	if err != nil {
		return nil, err
	}

	out := api.CorrelatedAssetGroupsResponse{
		GroupsByCorrelation: make(map[string][]api.CorrelatedAssetGroup),
	}

	for k, v := range assetGroups {
		key := fmt.Sprintf("%0.2f", k)
		out.GroupsByCorrelation[key] = []api.CorrelatedAssetGroup{}
		for _, ag := range v {
			out.GroupsByCorrelation[key] = append(out.GroupsByCorrelation[key], api.CorrelatedAssetGroup{
				Symbols:    ag.Symbols,
				TotalValue: ag.TotalValue,
			})
		}
	}

	return &out, nil
}
