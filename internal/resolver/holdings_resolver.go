package resolver

import (
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"

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
		return nil, err
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

	err = r.IngestionService.UpdatePosition(tx, req.TradingAccountID, domain.Position{
		Symbol:         req.Position.Symbol,
		Quantity:       decimal.NewFromFloat(req.Position.Quantity),
		TotalCostBasis: decimal.Zero, // should make explicit
	})
	if err != nil {
		return nil, err
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
