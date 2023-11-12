package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	"log"
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
	GetHoldings(accessToken string)
	GetTransactions(
		ctx context.Context,
		mappedTradingAccountIDs map[string]uuid.UUID,
		accessToken string,
	) ([]domain.Trade, error)
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

func (h plaidRepositoryHandler) GetHoldings(accessToken string) {
	ctx := context.Background()

	holdingsGetReq := plaid.NewInvestmentsHoldingsGetRequest(accessToken)

	// holdingsGetReqOptions := plaid.NewInvestmentHoldingsGetRequestOptions()
	// holdingsGetReqOptions.SetAccountIds([]string{"ACCOUNT_ID"})
	// holdingsGetReq.SetOptions(*holdingsGetReqOptions)

	resp, _, err := h.client.PlaidApi.InvestmentsHoldingsGet(ctx).InvestmentsHoldingsGetRequest(*holdingsGetReq).Execute()
	if err != nil {
		log.Fatal(err)
	}

	mappedSecurities := map[string]plaid.Security{}

	for _, s := range resp.Securities {
		mappedSecurities[s.SecurityId] = s
	}

	positionsByAccount := map[string][]domain.Position{}
	for _, holding := range resp.Holdings {
		if _, ok := positionsByAccount[holding.AccountId]; !ok {
			positionsByAccount[holding.AccountId] = []domain.Position{}
		}
		positionsByAccount[holding.AccountId] = append(positionsByAccount[holding.AccountId], domain.Position{
			Symbol:   *mappedSecurities[holding.SecurityId].TickerSymbol.Get(),
			Quantity: decimal.NewFromFloat32(holding.Quantity),
			// idk if this is true but it seems "holdings" compiles everything
			// into single position. if this isn't true, add another layer of mapps ig
			OpenLots: []domain.OpenLot{
				{
					Quantity: decimal.NewFromFloat32(holding.Quantity),
					// This field is calculated by Plaid as the sum of the purchase price of all of the shares in the holding.
					CostBasis: decimal.NewFromFloat32(*holding.CostBasis.Get()),
					Trade:     nil,        // aw hell nah man, wtf
					Date:      time.Now(), // bruh plaid doesnt have this field
				},
			},
		})
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(bytes))
}

func (h plaidRepositoryHandler) GetTransactions(
	ctx context.Context,
	mappedTradingAccountIDs map[string]uuid.UUID,
	accessToken string,
) ([]domain.Trade, error) {
	txGetRequest := plaid.NewInvestmentsTransactionsGetRequest(
		accessToken,
		time.Now().AddDate(-2, -1, 0).Format(time.DateOnly),
		time.Now().Format(time.DateOnly),
	)
	plaidAccountIDs := []string{}
	for k := range mappedTradingAccountIDs {
		plaidAccountIDs = append(plaidAccountIDs, k)
	}

	fmt.Println(plaidAccountIDs)

	opts := plaid.NewInvestmentsTransactionsGetRequestOptions()
	opts.SetAccountIds(plaidAccountIDs)
	txGetRequest.SetOptions(
		*opts,
	)

	resp, _, err := h.client.PlaidApi.InvestmentsTransactionsGet(ctx).InvestmentsTransactionsGetRequest(*txGetRequest).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get investment transactions from Plaid: %w", wrapPlaidError(err))
	}

	mappedSecurities := map[string]plaid.Security{}

	for _, s := range resp.Securities {
		securityType := *s.Type.Get()
		// seeing etf's with ticker symbol "NHX105509"
		if (strings.EqualFold(securityType, "etf") ||
			strings.EqualFold(securityType, "equity") ||
			strings.EqualFold(securityType, "mutual fund")) &&
			s.TickerSymbol.IsSet() && len(*s.TickerSymbol.Get()) < 6 {
			mappedSecurities[s.SecurityId] = s
		}
	}

	out := []domain.Trade{}
	for _, t := range resp.InvestmentTransactions {
		date, err := time.Parse(time.DateOnly, t.Date)
		if err != nil {
			return nil, err
		}
		security, ok := mappedSecurities[*t.SecurityId.Get()]
		if ok {
			// todo - handle dividends

			out = append(out, domain.Trade{
				Symbol:           *security.TickerSymbol.Get(),
				Quantity:         decimal.NewFromFloat32(t.Quantity).Abs(),
				Price:            decimal.NewFromFloat32(t.Price),
				Date:             date,
				Description:      &t.Name,
				TradingAccountID: mappedTradingAccountIDs[t.AccountId],
				Action:           model.TradeActionType(t.Subtype),
				IdempotencyKey:   t.InvestmentTransactionId,
			})
		}
	}
	return out, nil
}

func wrapPlaidError(err error) error {
	// conversionErr represnts an error converting err to PlaidError
	plaidErr, conversionErr := plaid.ToPlaidError(err)
	if conversionErr != nil {
		return fmt.Errorf("plaid_repository_error: %v. could not convert to Plaid error: %w", err, conversionErr)
	}
	return fmt.Errorf("plaid_repository_error %s: %s: %s", plaidErr.ErrorType, plaidErr.ErrorCode, plaidErr.ErrorMessage)
}
