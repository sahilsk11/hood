package resolver

import (
	"context"
	"database/sql"
	"hood/internal/repository"
	"hood/internal/service"

	api_types "github.com/sahilsk11/ace-common/types/hood"
)

type Resolver interface {
	// plaid endpoints
	GeneratePlaidLinkToken(ctx context.Context, req api_types.GeneratePlaidLinkTokenRequest) (*api_types.GeneratePlaidLinkTokenResponse, error)
	AddPlaidBankItem(ctx context.Context, req api_types.AddPlaidBankItemRequest) error

	// holdings endpoints
	NewManualTradingAccount(api_types.NewManualTradingAccountRequest) (*api_types.NewManualTradingAccountResponse, error)
	UpdatePosition(api_types.UpdatePositionRequest) (*api_types.UpdatePositionResponse, error)
	GetTradingAccountHoldings(req api_types.GetTradingAccountHoldingsRequest) (*api_types.GetTradingAccountHoldingsResponse, error)
	GetHistoricHoldings(req api_types.GetHistoricHoldingsRequest) (*api_types.GetHistoricHoldingsResponse, error)
}

type resolverHandler struct {
	Db                       *sql.DB
	PlaidRepository          repository.PlaidRepository
	UserRepository           repository.UserRepository
	PlaidItemRepository      repository.PlaidItemRepository
	TradingAccountRepository repository.TradingAccountRepository

	IngestionService service.IngestionService
	HoldingsService  service.HoldingsService
}

func NewResolver(
	db *sql.DB,
	plaidRepository repository.PlaidRepository,
	userRepository repository.UserRepository,
	plaidItemRepository repository.PlaidItemRepository,
	tradingAccountRepository repository.TradingAccountRepository,
	ingestionService service.IngestionService,
	holdingsService service.HoldingsService,
) Resolver {
	return resolverHandler{
		Db:                       db,
		PlaidRepository:          plaidRepository,
		UserRepository:           userRepository,
		PlaidItemRepository:      plaidItemRepository,
		TradingAccountRepository: tradingAccountRepository,
		IngestionService:         ingestionService,
		HoldingsService:          holdingsService,
	}
}
