package main

import (
	"hood/api"
	db "hood/internal/db/query"
	"hood/internal/resolver"
	"log"
)

func main() {
	dbConn, err := db.New()
	if err != nil {
		log.Fatal(err)
	}

	r := resolver.Resolver{
		Db: dbConn,
	}

	err = api.StartApi(5001, r)
	if err != nil {
		log.Fatal(err)
	}
}
