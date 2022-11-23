package main

import (
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

	Deps := data_ingestion.Deps{
		Db: db,
	}

	var cmd string
	flag.StringVar(&cmd, "command", "", "")
	flag.Parse()

	switch cmd {
	case "process-outfile":
		err = Deps.ProcessOutfile()
		if err != nil {
			log.Fatal(err)
		}
	case "update-prices":
		err = Deps.UpdatePrices()
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command '%s'", cmd)
	}

}
