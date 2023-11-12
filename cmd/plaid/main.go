package main

import (
	"context"
	"fmt"
	"hood/internal/repository"
	"hood/internal/util"
	"log"

	"github.com/google/uuid"
)

func main() {
	secrets, err := util.LoadSecrets()
	if err != nil {
		log.Fatal(err)
	}

	plaidRepository := repository.NewPlaidRepository(
		secrets.Plaid.ClientID,
		secrets.Plaid.Secret,
	)

	trades, err := plaidRepository.GetTransactions(
		context.Background(),
		map[string]uuid.UUID{
			"qrWRbqW3GpHkMmgXgJ7BtEPXbMn6myF6MmdZD": uuid.MustParse("1ea1069c-f711-4a39-9f3d-d95e476b28c5"),
		},
		"access-sandbox-6f16e8da-74a1-4cf4-b5b3-ada58cbf9ade",
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range trades {
		fmt.Println(t)
	}
}
