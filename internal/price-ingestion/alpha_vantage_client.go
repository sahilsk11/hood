package price_ingestion

import (
	"encoding/json"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/shopspring/decimal"
)

type AlphaVantageClient struct {
	httpClient http.Client
	apiKey     string
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
}

func (c AlphaVantageClient) GetLatestPrice(symbol string) (*model.Price, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, c.apiKey)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.httpClient.Do(req)
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
	// API uses odd format which includes numbers in JSON keys
	cleanedResponseBytes := cleanResponseBody(responseBytes)
	fmt.Println(string(cleanedResponseBytes))
	err = json.Unmarshal(cleanedResponseBytes, &responseJson)
	if err != nil {
		return nil, err
	}

	price, err := decimal.NewFromString(responseJson.GlobalQuote.Price)
	if err != nil {
		return nil, err
	}

	return &model.Price{
		Symbol:    responseJson.GlobalQuote.Symbol,
		Price:     price,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func cleanResponseBody(bytes []byte) []byte {
	r := regexp.MustCompile("\"[0-9]+\\. ")
	return r.ReplaceAll(bytes, []byte("\""))
}
