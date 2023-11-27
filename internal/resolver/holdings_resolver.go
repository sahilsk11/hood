package resolver

import (
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"

	"github.com/sahilsk11/ace-common/types/date"
	api_types "github.com/sahilsk11/ace-common/types/hood"
	"github.com/shopspring/decimal"
)

func (r resolverHandler) GetTradingAccountHoldings(req api_types.GetTradingAccountHoldingsRequest) (*api_types.GetTradingAccountHoldingsResponse, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	holdings, err := r.HoldingsService.GetCurrentPortfolio(tx, req.TradingAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current portfolio for %s: %w", req.TradingAccountID.String(), err)
	}

	out := []api_types.Position{}
	for _, p := range holdings.Positions {
		out = append(out, api_types.Position{
			Symbol:   p.Symbol,
			Quantity: p.Quantity.InexactFloat64(),
		})
	}

	return &api_types.GetTradingAccountHoldingsResponse{
		Positions: out,
		Cash:      holdings.Cash.InexactFloat64(),
	}, nil
}

func (r resolverHandler) UpdatePosition(req api_types.UpdatePositionRequest) (*api_types.UpdatePositionResponse, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, p := range req.Positions {
		err = r.IngestionService.UpdatePosition(tx, req.TradingAccountID, domain.Position{
			Symbol:         p.Symbol,
			Quantity:       decimal.NewFromFloat(p.Quantity),
			TotalCostBasis: decimal.Zero, // should make explicit
		})
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &api_types.UpdatePositionResponse{}, nil
}

func (r resolverHandler) NewManualTradingAccount(req api_types.NewManualTradingAccountRequest) (*api_types.NewManualTradingAccountResponse, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	acc, err := r.TradingAccountRepository.Add(tx, req.UserID, model.CustodianType_Unknown, model.AccountType_Unknown, nil, model.TradingAccountDataSourceType_Positions)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &api_types.NewManualTradingAccountResponse{
		TradingAccountID: acc.TradingAccountID,
	}, nil
}

func (r resolverHandler) GetHistoricHoldings(req api_types.GetHistoricHoldingsRequest) (*api_types.GetHistoricHoldingsResponse, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	results, err := r.HoldingsService.GetHistoricPortfolio(tx, req.TradingAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get historic portfolio: %w", err)
	}
	out := api_types.GetHistoricHoldingsResponse{
		Activity: make([]api_types.HoldingsActivity, len(*results)),
	}
	for i, r := range *results {
		positionsMap := r.Portfolio.ToHoldings().Positions
		positions := []domain.Position{}
		for _, p := range positionsMap {
			positions = append(positions, *p)
		}
		out.Activity[i] = api_types.HoldingsActivity{
			Positions:   positionsToApiPositions(positions),
			Date:        date.ProtoDateFromT(r.Date),
			Cash:        r.Portfolio.Cash.InexactFloat64(),
			Withdrawals: 0,
			Deposits:    0,
		}
	}

	return &out, nil
}

func positionsToApiPositions(in []domain.Position) []api_types.Position {
	out := make([]api_types.Position, len(in))
	for i, p := range in {
		out[i] = api_types.Position{
			Symbol:   p.Symbol,
			Quantity: p.Quantity.InexactFloat64(),
		}
	}
	return out
}
