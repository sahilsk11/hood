package types

type PortfolioCorrelationRequest struct {
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

type PortfolioCorrelationResponse struct {
	Correlations []Correlation `json:"correlations"`
}
