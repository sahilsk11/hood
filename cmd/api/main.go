package main

import (
	"hood/api"
	db "hood/internal/db/query"
	"hood/internal/repository"
	"hood/internal/resolver"
	"hood/internal/service"
	"hood/internal/util"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	env := util.Development
	secrets, err := util.LoadSecrets(env)
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
		env.ToPlaidEnv(),
	)

	userRepository := repository.NewUserRepository(dbConn)
	plaidItemRepository := repository.NewPlaidItemRepository(dbConn)
	tradingAccountRepository := repository.NewTradingAccountRepository(dbConn)
	tradeRepository := repository.NewTradeRepository()
	plaidInvestmentsAccountRepository := repository.NewPlaidInvestmentsHoldingsRepository(dbConn)

	ingestionService := service.NewIngestionService(
		plaidRepository,
		tradeRepository,
		plaidItemRepository,
		tradingAccountRepository,
		plaidInvestmentsAccountRepository,
	)
	holdingsService := service.NewHoldingsService(
		tradeRepository,
		tradingAccountRepository,
	)

	r := resolver.NewResolver(
		dbConn,
		plaidRepository,
		userRepository,
		plaidItemRepository,
		tradingAccountRepository,
		ingestionService,
		holdingsService,
	)

	err = api.StartApi(5001, r)
	if err != nil {
		log.Fatal(err)
	}
}
