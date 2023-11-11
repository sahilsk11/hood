package main

import (
	"hood/api"
	db "hood/internal/db/query"
	"hood/internal/repository"
	"hood/internal/resolver"
	"hood/internal/util"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	secrets, err := util.LoadSecrets()
	if err != nil {
		log.Fatal(err)
	}

	dbConn, err := db.New()
	if err != nil {
		log.Fatal(err)
	}

	plaidRepository := repository.NewPlaidRepository(
		secrets.Plaid.ClientID,
		secrets.Plaid.Secret,
	)

	userRepository := repository.NewUserRepository(dbConn)
	plaidItemRepository := repository.NewPlaidItemRepository(dbConn)
	tradingAccountRepository := repository.NewTradingAccountRepository(dbConn)

	r := resolver.Resolver{
		Db:                       dbConn,
		PlaidRepository:          plaidRepository,
		UserRepository:           userRepository,
		PlaidItemRepository:      plaidItemRepository,
		TradingAccountRepository: tradingAccountRepository,
	}

	err = api.StartApi(5001, r)
	if err != nil {
		log.Fatal(err)
	}
}
