package main

import (
	"fmt"
	"hood/internal/repository"
	"hood/internal/util"
	"log"

	"github.com/google/uuid"
)

func main() {
	env := util.Development
	secrets, err := util.LoadSecrets(env)
	if err != nil {
		log.Fatal(err)
	}

	plaidRepository := repository.NewPlaidRepository(
		secrets.Plaid.ClientID,
		secrets.Plaid.Secret,
		env.ToPlaidEnv(),
	)

	holdings, err := plaidRepository.GetHoldings(
		"access-sandbox-6f16e8da-74a1-4cf4-b5b3-ada58cbf9ade",
		map[string]uuid.UUID{
			"KqGn4yGoK9cJRKrvrzoXH6p4ozQk1vuENZVLP": uuid.MustParse("a836246d-8a8a-41a2-8f81-40d39a38059d"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(holdings)
}
