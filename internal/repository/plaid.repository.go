package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"hood/internal/domain"
	"log"

	"github.com/plaid/plaid-go/plaid"
	"github.com/shopspring/decimal"
)

type PlaidRepository interface {
	GetLinkToken(ctx context.Context) (string, error)
	GetAccessToken(publicToken string) error
	GetHoldings(accessToken string)
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

func (h plaidRepositoryHandler) GetLinkToken(ctx context.Context) (string, error) {
	var emailAddress string = "sk@gmail.com"
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: "e1607ec1-b9a1-452c-b966-8fe682e16a8a",
		EmailAddress: &emailAddress,
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

func (h plaidRepositoryHandler) GetAccessToken(publicToken string) error {
	ctx := context.Background()
	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangePublicTokenResp, _, err := h.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	if err != nil {
		return err
	}

	// These values should be saved to a persistent database and
	// associated with the currently signed-in user
	accessToken := exchangePublicTokenResp.GetAccessToken()
	itemID := exchangePublicTokenResp.GetItemId()

	fmt.Println("accessToken", accessToken)
	fmt.Println("itemID", itemID)

	return nil
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
					Trade:     nil, // aw hell nah man, wtf
					// Date: , // bruh plaid doesnt have this field
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
