package main

import (
	"context"
	"database/sql"
	"flag"
	data_ingestion "hood/internal/data-ingestion"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgresql://postgres:postgres@localhost:5438/postgres?sslmode=disable"
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

	var cmd string
	flag.StringVar(&cmd, "command", "", "")
	flag.Parse()

	switch cmd {
	case "process-outfile":
		err = data_ingestion.ProcessOutfile(ctx)
		if err != nil {
			log.Fatal(err)
		}
	case "update-prices":
		err = data_ingestion.UpdatePrices(ctx)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command '%s'", cmd)
	}
	tx.Commit()

}
