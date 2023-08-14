package main

import (
	"hood/internal/prices"
	"hood/internal/util"
	"log"
	"net/http"
	"time"
)

func main() {
	secrets, err := util.LoadSecrets()
	if err != nil {
		log.Fatal(err)
	}

	priceClient := prices.AlphaVantageClient{
		HttpClient: http.DefaultClient,
		ApiKey:     secrets.AlphaVantageKey,
	}

	out, err := priceClient.GetHistoricalPrices("AAPL", time.Now().AddDate(0, 0, -7))
	if err != nil {
		log.Fatal(err)
	}
	util.Pprint(out)
}
