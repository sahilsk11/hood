package main

import (
	db "hood/internal/db/query"
	"hood/internal/prices"
	"hood/internal/util"
	"log"
	"net/http"
)

func main() {
	tx, err := db.NewTx()
	if err != nil {
		log.Fatal(err)
	}

	secrets, err := util.LoadSecrets()
	if err != nil {
		log.Fatal(err)
	}

	djClient := prices.DataJockeyClient{
		HttpClient: http.DefaultClient,
		ApiKey:     secrets.DataJockeryApiKey,
	}

	err = prices.UpdateHistoricMetrics(tx, djClient, []string{"AAPL"})
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

}
