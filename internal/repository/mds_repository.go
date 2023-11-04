package repository

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	api "github.com/sahilsk11/ace-common/types/mds"
)

const mdsEndpoint = "http://localhost:5002"

func LatestPrices(req api.LatestPricesRequest) (*api.LatestPricesResponse, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(mdsEndpoint+"/latestPrices", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var resp api.LatestPricesResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func GetAdjustedPrices(req api.GetAdjustedPricesRequest) (*api.GetAdjustedPricesResponse, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(mdsEndpoint+"/adjustedPrices", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var resp api.GetAdjustedPricesResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func GetPricePercentChanges(req api.GetPricePercentChangesRequest) (*api.GetPricePercentChangesResponse, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(mdsEndpoint+"/pricePercentChanges", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var resp api.GetPricePercentChangesResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
