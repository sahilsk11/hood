package trade_ingestion

import (
	"context"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const fileContents = `
DOGE
5,306.22
$0.077957
+4.73%
AAPL
46.22 Shares
$150.18
+1.47%
TSLA
19.20 Shares
$169.98
+1.26%
GOOG
63.06 Shares
$97.33
+1.57%
NVDA
36.28 Shares
$160.38
+4.71%
ABNB
8.43 Shares
$94.42
-1.35%
UAL
8.20 Shares
$43.43
-0.00%
MSFT
4.01 Shares
$245.03
+1.23%
AMZN
10.48 Shares
$93.20
+0.80%
SQ
31.92 Shares
$62.40
+0.89%
CRM
4.52 Shares
$149.25
+3.04%
COIN
13.47 Shares
$42.80
+3.81%
SFIX
5.98 Shares
$3.70
-1.07%
DIS
7.55 Shares
$96.21
-1.40%
HOOD
2.72 Shares
$9.09
+2.71%
TEAM
0.681328 Shares
$116.34
-1.04%
`

var tickerMap = map[string]string{
	"SIRIUS XM":                        "SIRI",
	"AMAG PHARMACEUTICALS":             "AMAG",
	"SONOS":                            "SONO",
	"GAP":                              "GPS",
	"SNAP":                             "SNAP",
	"SPOTIFY":                          "SPOT",
	"BITCOIN":                          "BTC",
	"UBER":                             "UBER",
	"META PLATFORMS":                   "META",
	"TWITTER":                          "TWTR",
	"APPLE":                            "AAPL",
	"TESLA":                            "TSLA",
	"ALPHABET CLASS A":                 "GOOGL",
	"ALPHABET CLASS C":                 "GOOG",
	"NVIDIA":                           "NVDA",
	"UNITED AIRLINES":                  "UAL",
	"MICROSOFT":                        "MSFT",
	"AMAZON":                           "AMZN",
	"BLOCK":                            "SQ",
	"SALESFORCE":                       "CRM",
	"COINBASE":                         "COIN",
	"STITCH FIX":                       "SFIX",
	"DISNEY":                           "DIS",
	"ROBINHOOD MARKETS":                "HOOD",
	"ATLASSIAN":                        "TEAM",
	"ATLASSIAN CORPORATION":            "TEAM",
	"DOGECOIN":                         "DOGE",
	"AIRBNB":                           "ABNB",
	"SAP":                              "SAP",
	"DATADOG":                          "DDOG",
	"SPDR GOLD TRUST":                  "GLD",
	"VANGUARD HIGH DIVIDEND YIELD ETF": "VYM",
	"SPDR PORTFOLIO S&P 500 HIGH DIVIDEND ETF": "SPYD",
	"SIMON PROPERTY GROUP":                     "SPG",
	"NIO":                                      "NIO",
	"ETHEREUM":                                 "ETH",
	"GAMESTOP":                                 "GME",
	"AMC ENTERTAINMENT":                        "AMC",
	"DOORDASH":                                 "DASH",
	"SPDR S&P 500 ETF":                         "SPY",
	"XILINX":                                   "XLNX",
}

func tickerToName(ticker string) (string, error) {
	for k, v := range tickerMap {
		if v == ticker {
			return k, nil
		}
	}

	return "", fmt.Errorf("could not map ticker '%s' to name", ticker)
}

func nameToTicker(name string) (string, error) {
	ticker, ok := tickerMap[name]
	if !ok {
		return "", fmt.Errorf("could not map name '%s' to ticker", name)
	}

	return ticker, nil
}

func parseRhPrices(textExport string) ([]model.Price, error) {
	textExport = strings.Trim(textExport, " \n\t")
	lines := strings.Split(textExport, "\n")

	tickerRegex, _ := regexp.Compile("^[a-zA-Z]+$")
	pricesToUpdate := []model.Price{}

	i := 0
	for i < len(lines) {
		line := lines[i]
		if tickerRegex.MatchString(line) {
			ticker := line
			// search for next line starting with $
			for string(lines[i][0]) != "$" {
				i++
			}
			priceStr := strings.ReplaceAll(lines[i], "$", "")
			priceStr = strings.ReplaceAll(priceStr, ",", "")

			price, err := decimal.NewFromString(priceStr)
			if err != nil {
				return nil, fmt.Errorf("could parse price from '%s': %w", priceStr, err)
			}
			// name, err := tickerToName(ticker)
			// if err != nil {
			// 	return nil, err
			// }
			pricesToUpdate = append(pricesToUpdate, model.Price{
				Symbol:    ticker,
				Price:     price,
				UpdatedAt: time.Now().UTC(),
			})
		}
		i++
	}

	return pricesToUpdate, nil
}

func UpdatePrices(ctx context.Context) error {
	prices, err := parseRhPrices(fileContents)
	if err != nil {
		return fmt.Errorf("failed to parse prices file: %w", err)
	}
	_, err = db.AddPricesToDb(ctx, prices)
	if err != nil {
		return err
	}

	return nil
}
