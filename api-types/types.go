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

type CorrelatedAssetGroupsRequest struct {
	Holdings []HoldingsPosition `json:"holdings"`
}

type CorrelatedAssetGroup struct {
	Symbols    []string `json:"symbols"`
	TotalValue float64  `json:"totalValue"`
}
type CorrelatedAssetGroupsResponse struct {
	GroupsByCorrelation map[string][]CorrelatedAssetGroup `json:"groupsByCorrelation"`
}
