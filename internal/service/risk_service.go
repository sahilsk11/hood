package service

import (
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/metrics"
	"hood/internal/util"
	"sort"
	"strings"

	"time"
)

// https://icfs.com/financial-knowledge-center/importance-standard-deviation-investment#:~:text=With%20most%20investments%2C%20including%20mutual,standard%20deviation%20would%20be%20zero.
const stdevRange = 3 * (time.Hour * 24 * 365)

type AssetCorrelation struct {
	AssetOne    string
	AssetTwo    string
	Correlation float64
}

func PortfolioCorrelation(tx *sql.Tx, symbols []string) ([]AssetCorrelation, error) {
	start := time.Now().Add(-1 * stdevRange)
	prices, err := db.GetAdjustedPrices(tx, symbols, start)
	if err != nil {
		return nil, err
	}

	return calculatePortfolioCorrelationWithPrices(prices, symbols)
}

func calculatePortfolioCorrelationWithPrices(prices []model.Price, symbols []string) ([]AssetCorrelation, error) {
	dailyPercentChanges, err := metrics.CalculateDailyPercentChange(prices)
	if err != nil {
		return nil, err
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

type nodeNeighbor struct {
	Node   node
	Length float64
}
type node struct {
	Name      string
	Neighbors map[string]nodeNeighbor
}

func CalculateCorrelationAllocation(tx *sql.Tx, portfolio domain.MetricsPortfolio) ([]CorrelationAllocation, error) {
	// start := time.Now().Add(-1 * stdevRange)
	// prices, err := db.GetAdjustedPrices(tx, portfolio.Symbols(), start)
	// if err != nil {
	// 	return nil, err
	// }

	// correlations, err := calculatePortfolioCorrelationWithPrices(prices, portfolio.Symbols())
	// if err != nil {
	// 	return nil, err
	// }

	// nodes := correl  /ationsToGraph(correlations)
	return nil, nil
}

func correlationsToGraph(correlations []AssetCorrelation) []node {
	mappedNodes := map[string]node{}
	for _, c := range correlations {
		// add to node map
		{
			if _, ok := mappedNodes[c.AssetOne]; !ok {
				mappedNodes[c.AssetOne] = node{
					Name:      c.AssetOne,
					Neighbors: map[string]nodeNeighbor{},
				}
			}
			if _, ok := mappedNodes[c.AssetTwo]; !ok {
				mappedNodes[c.AssetTwo] = node{
					Name:      c.AssetTwo,
					Neighbors: map[string]nodeNeighbor{},
				}
			}
		}

		mappedNodes[c.AssetOne].Neighbors[c.AssetTwo] = nodeNeighbor{
			Node:   mappedNodes[c.AssetTwo],
			Length: c.Correlation,
		}
		mappedNodes[c.AssetTwo].Neighbors[c.AssetOne] = nodeNeighbor{
			Node:   mappedNodes[c.AssetOne],
			Length: c.Correlation,
		}
	}

	nodes := []node{}
	for _, v := range mappedNodes {
		nodes = append(nodes, v)
	}

	return nodes
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

func keepLargestGroups(groups [][]string, valueBySymbol map[string]float64) [][]string {
	// if a symbol exists in more than one group,
	// only keep the group which has the highest value
	used := util.NewSet()
	type node struct {
		symbols []string
		value   float64
	}
	nodes := []node{}
	for _, group := range groups {
		value := 0.0
		for _, symbol := range group {
			value += valueBySymbol[symbol]
		}
		nodes = append(nodes, node{
			symbols: group,
			value:   value,
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].value > nodes[j].value
	})
	out := [][]string{}
	for _, node := range nodes {
		groupUsed := false
		for _, symbol := range node.symbols {
			if !used.Contains(symbol) {
				used.Add(symbol)
			} else {
				groupUsed = true
			}
		}
		if !groupUsed {
			out = append(out, node.symbols)
		}
	}

	return out
}
