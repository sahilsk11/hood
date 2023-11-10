package main

import (
	db "hood/internal/db/query"

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
	defer tx.Rollback()

	tx.Commit()
}
