package service

import (
	"database/sql"
	"fmt"
	"hood/internal/domain"
	"hood/internal/metrics"
	"hood/internal/repository"
	"hood/internal/util"
	"sort"
	"strings"

	"time"

	"github.com/sahilsk11/ace-common/types/date"
	api "github.com/sahilsk11/ace-common/types/mds"
)

// https://icfs.com/financial-knowledge-center/importance-standard-deviation-investment#:~:text=With%20most%20investments%2C%20including%20mutual,standard%20deviation%20would%20be%20zero.
const stdevRange = 3 * (time.Hour * 24 * 365)

type AssetCorrelation struct {
	AssetOne    string
	AssetTwo    string
	Correlation float64
}

func PortfolioCorrelation(tx *sql.Tx, symbols []string) ([]AssetCorrelation, error) {
	resp, err := repository.GetPricePercentChanges(api.GetPricePercentChangesRequest{
		Symbols: symbols,
		Start:   date.ProtoDateFromT(time.Now().UTC().Add(-stdevRange)),
		End:     date.ProtoDateFromT(time.Now().UTC()),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.MissingDataSymbols) > 0 {
		return nil, fmt.Errorf("cannot calculate correlation - missing price data for %v", resp.MissingDataSymbols)
	}

	mappedPercentChanges := map[string]domain.PercentData{}
	for k, v := range resp.PercentChangesBySymbol {
		mappedPercentChanges[k] = domain.PercentData{}
		for _, x := range v {
			mappedPercentChanges[k] = append(mappedPercentChanges[k], domain.Percent(x.PercentChange))
		}
	}

	return calculatePortfolioCorrelationWithPrices(mappedPercentChanges)
}

func calculatePortfolioCorrelationWithPrices(dailyPercentChanges map[string]domain.PercentData) ([]AssetCorrelation, error) {
	symbols := []string{}
	for k := range dailyPercentChanges {
		symbols = append(symbols, k)
	}

	sort.Strings(symbols)

	out := []AssetCorrelation{}
	for i, s1 := range symbols {
		for j := i + 1; j < len(symbols); j++ {
			s2 := symbols[j]

			// there appear to be some duplicate values n shit. the inputs should actually be cleaned
			// better so days line up and we do best approx with days we have. for now this is ok.
			firstN := len(dailyPercentChanges[s1])
			if len(dailyPercentChanges[s2]) < firstN {
				firstN = len(dailyPercentChanges[s2])
			}

			corr, err := metrics.Correlation(
				dailyPercentChanges[s1][:firstN],
				dailyPercentChanges[s2][:firstN],
			)
			if err != nil {
				return nil, err
			}

			out = append(out, AssetCorrelation{
				AssetOne:    s1,
				AssetTwo:    s2,
				Correlation: corr,
			})
		}
	}

	return out, nil
}

type CorrelationAllocation struct {
	Correlation   float64
	ValueBySymbol map[string]float64
}

type CorrelatedAssetGroup struct {
	Symbols    []string
	TotalValue float64
}

func CalculateCorrelatedAssetGroups(tx *sql.Tx, portfolio domain.MetricsPortfolio) (map[float64][]CorrelatedAssetGroup, error) {
	priceChangesResponse, err := repository.GetPricePercentChanges(api.GetPricePercentChangesRequest{
		Symbols: portfolio.Symbols(),
		Start:   date.ProtoDateFromT(time.Now().UTC().Add(-stdevRange)),
		End:     date.ProtoDateFromT(time.Now().UTC()),
	})
	if err != nil {
		return nil, err
	}
	if len(priceChangesResponse.MissingDataSymbols) > 0 {
		return nil, fmt.Errorf("cannot calculated correlated asset groups - missing price data for %v", priceChangesResponse.MissingDataSymbols)
	}

	latestSymbolsResponse, err := repository.LatestPrices(api.LatestPricesRequest{
		Symbols: portfolio.Symbols(),
	})
	if err != nil {
		return nil, err
	}
	if len(latestSymbolsResponse.MissingDataSymbols) > 0 {
		return nil, fmt.Errorf("cannot calculate correlated asset groups - missing latest price for %v", latestSymbolsResponse.MissingDataSymbols)
	}
	latestPrices := latestSymbolsResponse.PricesBySymbol

	mappedPercentChanges := map[string]domain.PercentData{}
	for k, v := range priceChangesResponse.PercentChangesBySymbol {
		mappedPercentChanges[k] = domain.PercentData{}
		for _, x := range v {
			mappedPercentChanges[k] = append(mappedPercentChanges[k], domain.Percent(x.PercentChange))
		}
	}

	correlations, err := calculatePortfolioCorrelationWithPrices(mappedPercentChanges)
	if err != nil {
		return nil, err
	}

	valueBySymbol := map[string]float64{}
	for _, p := range portfolio.Positions {
		valueBySymbol[p.Symbol] = latestPrices[p.Symbol] * p.Quantity.InexactFloat64()
	}

	out := map[float64][]CorrelatedAssetGroup{}
	for t := 0.0; t < 1.0; t += 0.1 {
		out[t] = groupCorrelatedAssets(correlations, t, valueBySymbol)
	}

	return out, nil
}

func groupCorrelatedAssets(correlations []AssetCorrelation, threshold float64, valueBySymbol map[string]float64) []CorrelatedAssetGroup {
	graph := correlationsToGraph(correlations)
	groupsWithDuplicateSymbols := createGroupsWithThreshold(graph, threshold)

	return keepLargestGroups(groupsWithDuplicateSymbols, valueBySymbol)
}

func correlationsToGraph(correlations []AssetCorrelation) graph {
	graph := graph{}
	for _, c := range correlations {
		if _, ok := graph[c.AssetOne]; !ok {
			graph[c.AssetOne] = map[string]float64{}
		}
		if _, ok := graph[c.AssetTwo]; !ok {
			graph[c.AssetTwo] = map[string]float64{}
		}

		graph[c.AssetOne][c.AssetTwo] = c.Correlation
		graph[c.AssetTwo][c.AssetOne] = c.Correlation
	}

	return graph
}

type graph map[string]map[string]float64

func dfs(node string, graph graph, visited *util.Set, currentGroup *util.Set, threshold float64) {
	visited.Add(node)
	currentGroup.Add(node)

	for neighbor := range graph[node] {
		if !visited.Contains(neighbor) {
			ok := true
			for neighorName, length := range graph[neighbor] {
				if currentGroup.Contains(neighorName) && length < threshold {
					ok = false
				}
			}
			if ok {
				dfs(neighbor, graph, visited, currentGroup, threshold)
			}
		}
	}
}

func createGroupsWithThreshold(graph graph, correlationThreshold float64) [][]string {
	groups := util.NewSet()
	for node := range graph {
		group := util.NewSet()
		visited := util.NewSet()
		dfs(node, graph, visited, group, correlationThreshold)
		groups.Add(strings.Join(group.List(), ","))
	}

	out := [][]string{}
	for _, g := range groups.List() {
		out = append(out, strings.Split(g, ","))
	}

	return out
}

func keepLargestGroups(groups [][]string, valueBySymbol map[string]float64) []CorrelatedAssetGroup {
	// if a symbol exists in more than one group,
	// only keep the group which has the highest value
	used := util.NewSet()
	assetGroups := []CorrelatedAssetGroup{}
	for _, group := range groups {
		value := 0.0
		for _, symbol := range group {
			value += valueBySymbol[symbol]
		}
		// ensure groups have more than one element
		if len(group) > 1 {
			assetGroups = append(assetGroups, CorrelatedAssetGroup{
				Symbols:    group,
				TotalValue: value,
			})
		}
	}
	sort.Slice(assetGroups, func(i, j int) bool {
		return assetGroups[i].TotalValue > assetGroups[j].TotalValue
	})
	out := []CorrelatedAssetGroup{}
	for _, ag := range assetGroups {
		groupUsed := false
		for _, symbol := range ag.Symbols {
			if !used.Contains(symbol) {
				used.Add(symbol)
			} else {
				groupUsed = true
			}
		}
		if !groupUsed {
			out = append(out, ag)
		}
	}

	return out
}
