package repository

import (
	"context"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/plaid/plaid-go/plaid"
	"github.com/shopspring/decimal"
)

type PlaidRepository interface {
	GetLinkToken(ctx context.Context, userID uuid.UUID, email string) (string, error)
	GetAccessToken(publicToken string) (
		accessToken string,
		itemID string,
		err error,
	)
	GetHoldings(
		accessToken string,
		mappedTradingAccountIDs map[string]uuid.UUID,
	) ([]model.PlaidInvestmentHoldings, error)
	GetTransactions(
		ctx context.Context,
		mappedTradingAccountIDs map[string]uuid.UUID,
		accessToken string,
	) ([]domain.Trade, []model.PlaidTradeMetadata, error)
}

type plaidRepositoryHandler struct {
	client *plaid.APIClient
}

func NewPlaidRepository(clientID, secret string) PlaidRepository {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(configuration)

	return plaidRepositoryHandler{
		client: client,
	}
}

func (h plaidRepositoryHandler) GetLinkToken(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userID.String(),
		EmailAddress: &email,
	}

	request := plaid.NewLinkTokenCreateRequest(
		"Investment Tracker",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
		user,
	)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_INVESTMENTS})
	request.SetWebhook("https://factorbacktest.net/plaidWebhook")
	// request.SetRedirectUri("https://domainname.com/oauth-page.html")

	linkTokenCreateResp, _, err := h.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return "", err
	}

	return linkTokenCreateResp.LinkToken, nil
}

func (h plaidRepositoryHandler) GetAccessToken(publicToken string) (
	accessToken string,
	itemID string,
	err error,
) {
	ctx := context.Background()
	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangePublicTokenResp, _, err := h.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	if err != nil {
		return "", "", err
	}

	// These values should be saved to a persistent database and
	// associated with the currently signed-in user
	accessToken = exchangePublicTokenResp.GetAccessToken()
	itemID = exchangePublicTokenResp.GetItemId()

	// if we fail during this call, we will have ghost item ids

	return accessToken, itemID, nil
}

func (h plaidRepositoryHandler) GetHoldings(
	accessToken string,
	mappedTradingAccountIDs map[string]uuid.UUID,
) ([]model.PlaidInvestmentHoldings, error) {
	ctx := context.Background()

	plaidAccountIDs := []string{}
	for k := range mappedTradingAccountIDs {
		plaidAccountIDs = append(plaidAccountIDs, k)
	}

	holdingsGetReq := plaid.NewInvestmentsHoldingsGetRequest(accessToken)
	holdingsGetReq.Options = &plaid.InvestmentHoldingsGetRequestOptions{
		AccountIds: &plaidAccountIDs,
	}

	resp, _, err := h.client.PlaidApi.InvestmentsHoldingsGet(ctx).InvestmentsHoldingsGetRequest(*holdingsGetReq).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get investment holdings from Plaid: %w", wrapPlaidError(err))
	}

	mappedSecurities := filterSecurities(resp.Securities)
	out := []model.PlaidInvestmentHoldings{}
	for _, holding := range resp.Holdings {
		if security, ok := mappedSecurities[holding.SecurityId]; ok {
			out = append(out, model.PlaidInvestmentHoldings{
				// PlaidInvestmentsHoldingsID: ,
				Ticker:           *security.TickerSymbol.Get(),
				TradingAccountID: mappedTradingAccountIDs[holding.GetAccountId()],
				TotalCostBasis:   decimal.NewFromFloat32(*holding.CostBasis.Get()),
				Quantity:         decimal.NewFromFloat32(holding.Quantity),
			})
		}
	}

	return out, nil
}

func (h plaidRepositoryHandler) GetTransactions(
	ctx context.Context,
	mappedTradingAccountIDs map[string]uuid.UUID,
	accessToken string,
) ([]domain.Trade, []model.PlaidTradeMetadata, error) {
	txGetRequest := plaid.NewInvestmentsTransactionsGetRequest(
		accessToken,
		time.Now().AddDate(-2, -1, 0).Format(time.DateOnly),
		time.Now().Format(time.DateOnly),
	)
	plaidAccountIDs := []string{}
	for k := range mappedTradingAccountIDs {
		plaidAccountIDs = append(plaidAccountIDs, k)
	}

	opts := plaid.NewInvestmentsTransactionsGetRequestOptions()
	opts.SetAccountIds(plaidAccountIDs)
	txGetRequest.SetOptions(
		*opts,
	)

	resp, _, err := h.client.PlaidApi.InvestmentsTransactionsGet(ctx).InvestmentsTransactionsGetRequest(*txGetRequest).Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get investment transactions from Plaid: %w", wrapPlaidError(err))
	}

	mappedSecurities := filterSecurities(resp.Securities)

	trades := []domain.Trade{}
	plaidTrades := []model.PlaidTradeMetadata{}

	for _, t := range resp.InvestmentTransactions {
		date, err := time.Parse(time.DateOnly, t.Date)
		if err != nil {
			return nil, nil, err
		}
		security, ok := mappedSecurities[*t.SecurityId.Get()]
		if ok {
			// todo - handle dividends

			trades = append(trades, domain.Trade{
				Symbol:           *security.TickerSymbol.Get(),
				Quantity:         decimal.NewFromFloat32(t.Quantity).Abs(),
				Price:            decimal.NewFromFloat32(t.Price),
				Date:             date,
				Description:      &t.Name,
				TradingAccountID: mappedTradingAccountIDs[t.AccountId],
				Action:           model.TradeActionType(t.Subtype),
				Source:           model.TradeSourceType_Plaid,
			})

			plaidTrades = append(plaidTrades, model.PlaidTradeMetadata{
				// PlaidTradeMetadataID: , // generated by db
				// TradeID:                      nil, // we don't know this yet. should switch to uuids so we can generate here
				PlaidInvestmentTransactionID: t.InvestmentTransactionId,
			})
		}
	}
	return trades, plaidTrades, nil
}

func filterSecurities(securities []plaid.Security) map[string]plaid.Security {
	out := map[string]plaid.Security{}
	for _, s := range securities {
		securityType := *s.Type.Get()
		// seeing etf's with ticker symbol "NHX105509"
		if (strings.EqualFold(securityType, "etf") ||
			strings.EqualFold(securityType, "equity") ||
			strings.EqualFold(securityType, "mutual fund")) &&
			s.TickerSymbol.IsSet() && s.TickerSymbol.Get() != nil && len(*s.TickerSymbol.Get()) < 6 {
			out[s.SecurityId] = s
		}
	}
	return out
}

func wrapPlaidError(err error) error {
	// conversionErr represnts an error converting err to PlaidError
	plaidErr, conversionErr := plaid.ToPlaidError(err)
	if conversionErr != nil {
		return fmt.Errorf("plaid_repository_error: %v. could not convert to Plaid error: %w", err, conversionErr)
	}
	return fmt.Errorf("plaid_repository_error %s: %s: %s", plaidErr.ErrorType, plaidErr.ErrorCode, plaidErr.ErrorMessage)
}
