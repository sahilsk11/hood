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
}

type resolverHandler struct {
	Db                       *sql.DB
	PlaidRepository          repository.PlaidRepository
	UserRepository           repository.UserRepository
	PlaidItemRepository      repository.PlaidItemRepository
	TradingAccountRepository repository.TradingAccountRepository

	IngestionService service.IngestionService
}

func NewResolver(
	db *sql.DB,
	plaidRepository repository.PlaidRepository,
	userRepository repository.UserRepository,
	plaidItemRepository repository.PlaidItemRepository,
	tradingAccountRepository repository.TradingAccountRepository,
	ingestionService service.IngestionService,
) Resolver {
	return resolverHandler{
		Db:                       db,
		PlaidRepository:          plaidRepository,
		UserRepository:           userRepository,
		PlaidItemRepository:      plaidItemRepository,
		TradingAccountRepository: tradingAccountRepository,
		IngestionService:         ingestionService,
	}
}
