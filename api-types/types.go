package types

type CorrelationMatrixRequest struct {
	// eventually will make this portfolio ID
	UseMyPortfolio bool `json:"useMyPortfolio"`
	// optional - use only if flag is false
	Symbols []string `json:"symbols"`
}

type Correlation struct {
	AssetOne    string  `json:"assetOne"`
	AssetTwo    string  `json:"assetTwo"`
	Correlation float64 `json:"correlation"`
}

type CorrelationMatrixResponse struct {
	Correlations []Correlation `json:"correlations"`
}

type HoldingsPosition struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
}

type CorrelationAllocationRequest struct {
	Holdings []HoldingsPosition `json:"holdings"`
}

type CorrelationAllocationResponse struct {
}
