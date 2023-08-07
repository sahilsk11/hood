package main

import (
	"fmt"
	db "hood/internal/db/query"
	"hood/internal/metrics"
	"log"
)

func main() {
	dbConn, err := db.New()
	if err != nil {
		log.Fatal(err)
	}
	tx, err := dbConn.Begin()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(metrics.MomentumFactorForAsset(tx, "TSLA"))

}
