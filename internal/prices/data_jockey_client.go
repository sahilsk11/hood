package prices

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DataJockeyClient struct {
	HttpClient *http.Client
	ApiKey     string
}

type DataJockeyFinancialResponse struct {
	Currency    string `json:"currency"`
	CompanyInfo struct {
		CIK    string `json:"cik"`
		Ticker string `json:"ticker"`
		Name   string `json:"name"`
	} `json:"company_info"`
	FinancialData struct {
		Quarterly struct {
			Revenue                             map[string]int64 `json:"revenue"`
			CostOfRevenue                       map[string]int64 `json:"cost_of_revenue"`
			GrossProfit                         map[string]int64 `json:"gross_profit"`
			OperatingIncome                     map[string]int64 `json:"operating_income"`
			TotalAssets                         map[string]int64 `json:"total_assets"`
			TotalCurrentAssets                  map[string]int64 `json:"total_current_assets"`
			PrepaidExpenses                     map[string]int64 `json:"prepaid_expenses"`
			PropertyPlantAndEquipmentNet        map[string]int64 `json:"property_plant_and_equipment_net"`
			RetainedEarnings                    map[string]int64 `json:"retained_earnings"`
			OtherAssetsNoncurrent               map[string]int64 `json:"other_assets_noncurrent"`
			TotalLiabilities                    map[string]int64 `json:"total_liabilities"`
			ShareholderEquity                   map[string]int64 `json:"shareholder_equity"`
			NetIncome                           map[string]int64 `json:"net_income"`
			OperatingCashFlow                   map[string]int64 `json:"operating_cash_flow"`
			InvestingCashFlow                   map[string]int64 `json:"investing_cash_flow"`
			FinancingCashFlow                   map[string]int64 `json:"financing_cash_flow"`
			ResearchDevelopmentExpense          map[string]int64 `json:"research_development_expense"`
			SellingGeneralAdministrativeExpense map[string]int64 `json:"selling_general_administrative_expense"`
			OperatingExpenses                   map[string]int64 `json:"operating_expenses"`
			NonOperatingIncome                  map[string]int64 `json:"non_operating_income"`
			PreTaxIncome                        map[string]int64 `json:"pre_tax_income"`
			IncomeTax                           map[string]int64 `json:"income_tax"`
			DepreciationAmortization            map[string]int64 `json:"depreciation_amortization"`
			StockBasedCompensation              map[string]int64 `json:"stock_based_compensation"`
			DividendsPaid                       map[string]int64 `json:"dividends_paid"`
			CashOnHand                          map[string]int64 `json:"cash_on_hand"`
			CurrentNetReceivables               map[string]int64 `json:"current_net_receivables"`
			Inventory                           map[string]int64 `json:"inventory"`
			TotalCurrentLiabilities             map[string]int64 `json:"total_current_liabilities"`
			TotalNonCurrentLiabilities          map[string]int64 `json:"total_non_current_liabilities"`
			LongTermDebt                        map[string]int64 `json:"long_term_debt"`
			Goodwill                            map[string]int64 `json:"goodwill"`
			IntangibleAssetsExcludingGoodwill   map[string]int64 `json:"intangible_assets_excluding_goodwill"`
		} `json:"quarterly"`
	} `json:"financial_data"`
}

func (c DataJockeyClient) GetAssetMetrics(symbol string) (*DataJockeyFinancialResponse, error) {
	url := fmt.Sprintf("https://api.datajockey.io/v0/company/financials?apikey=%s&ticker=%s&period=Q", c.ApiKey, symbol)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		if err != nil {
			return nil, err
		}
	}

	var responseJson DataJockeyFinancialResponse
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(responseBytes, &responseJson)
	if err != nil {
		return nil, err
	}

	return &responseJson, nil
}
