package prices

import (
	"encoding/json"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type AlphaVantageClient struct {
	HttpClient *http.Client
	ApiKey     string
}

type alphaVantageQuoteResult struct {
	GlobalQuote struct {
		Symbol           string `json:"symbol"`
		Open             string `json:"open"`
		High             string `json:"high"`
		Low              string `json:"low"`
		Price            string `json:"price"`
		Volume           string `json:"volume"`
		LatestTradingDay string `json:"latest trading day"`
		PreviousClose    string `json:"previous close"`
		Change           string `json:"change"`
		ChangePercent    string `json:"change percent"`
	} `json:"Global Quote"`
	Note string `json:"Note"`
}

type alphaVantageHistoricPriceResult struct {
	Metadata   struct{} `json:"Meta Data"`
	TimeSeries map[string]struct {
		Open  decimal.Decimal `json:"open"`
		Close decimal.Decimal `json:"close"`
	} `json:"Time Series (Daily)"`
	Note string `json:"Note"`
}

func (c AlphaVantageClient) GetLatestPrice(symbol string) (*model.Price, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, c.ApiKey)
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

	var responseJson alphaVantageQuoteResult
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(cleanResponseBody(responseBytes)))

	// API uses odd format which includes numbers in JSON keys
	cleanedResponseBytes := cleanResponseBody(responseBytes)
	err = json.Unmarshal(cleanedResponseBytes, &responseJson)
	if err != nil {
		return nil, err
	}

	if strings.Contains(responseJson.Note, "Our standard API call frequency is 5 calls per minute") {
		fmt.Println("alpha vantage rate limit hit, waiting")
		time.Sleep(time.Minute)
		return c.GetLatestPrice(symbol)
	}

	price, err := decimal.NewFromString(responseJson.GlobalQuote.Price)
	if err != nil {
		return nil, err
	}

	latestTradingDay, err := time.Parse("2006-01-02", responseJson.GlobalQuote.LatestTradingDay)
	if err != nil {
		return nil, fmt.Errorf("could not parse latest trading day from Alpha Vantage response: %w", err)
	}

	return &model.Price{
		Symbol:    responseJson.GlobalQuote.Symbol,
		Price:     price,
		Date:      latestTradingDay,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (c AlphaVantageClient) GetHistoricalPrices(symbol string, start time.Time) ([]model.Price, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s&outputsize=full", symbol, c.ApiKey)
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

	var responseJson alphaVantageHistoricPriceResult
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// API uses odd format which includes numbers in JSON keys
	cleanedResponseBytes := cleanResponseBody(responseBytes)
	err = json.Unmarshal(cleanedResponseBytes, &responseJson)
	if err != nil {
		return nil, err
	}

	if strings.Contains(responseJson.Note, "Our standard API call frequency is 5 calls per minute") {
		fmt.Println("alpha vantage rate limit hit, waiting")
		time.Sleep(time.Minute)
		return c.GetHistoricalPrices(symbol, start)
	}

	out := []model.Price{}
	for k, v := range responseJson.TimeSeries {
		d, err := time.Parse("2006-01-02", k)
		if err != nil {
			return nil, fmt.Errorf("could not parse price date of %s: %w", symbol, err)
		}
		if d.Unix() >= start.Unix() {
			out = append(out, model.Price{
				Symbol:    symbol,
				Price:     v.Close,
				UpdatedAt: time.Now(),
				Date:      d,
			})
		}
	}

	return out, nil
}

func cleanResponseBody(bytes []byte) []byte {
	r := regexp.MustCompile("\"[0-9]+\\. ")
	return r.ReplaceAll(bytes, []byte("\""))
}
