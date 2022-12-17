package main

import (
	"context"
	"database/sql"
	"fmt"
	trade_ingestion "hood/internal/trade-ingestion"
	"hood/internal/util"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	secrets, err := util.LoadSecrets()
	if err != nil {
		log.Fatal(err)
	}

	connStr := fmt.Sprintf(
		"postgresql://postgres:%s@hood-db-test.cp1ikxt0og0j.us-east-1.rds.amazonaws.com:5432/postgres?sslmode=disable",
		secrets.RdsPassword,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.WithValue(
		context.Background(),
		"tx",
		tx,
	)

	err = trade_ingestion.ProcessOutfile(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

}
