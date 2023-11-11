package resolver

import (
	"context"
	"fmt"
	"hood/internal/db/models/postgres/public/model"

	api_types "github.com/sahilsk11/ace-common/types/hood"
)

func (r Resolver) GeneratePlaidLinkToken(ctx context.Context, req api_types.GeneratePlaidLinkTokenRequest) (*api_types.GeneratePlaidLinkTokenResponse, error) {
	user, err := r.UserRepository.Get(req.UserID)
	if err != nil {
		return nil, err
	}

	linkToken, err := r.PlaidRepository.GetLinkToken(ctx, req.UserID, user.PrimaryEmail)
	if err != nil {
		return nil, err
	}

	return &api_types.GeneratePlaidLinkTokenResponse{
		LinkToken: linkToken,
	}, nil
}

func (r Resolver) AddPlaidBankItem(ctx context.Context, req api_types.AddPlaidBankItemRequest) error {
	tx, err := r.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	accessToken, itemID, err := r.PlaidRepository.GetAccessToken(req.PublicToken)
	if err != nil {
		return err
	}

	item, err := r.PlaidItemRepository.Add(tx, req.UserID, itemID, accessToken)
	if err != nil {
		return err
	}

	for _, acc := range req.Accounts {
		if acc.Type == "investment" {
			accountType := accountTypeFromPlaid(acc.Subtype)
			institution := institutionFromPlaid(req.InstitutionName)
			tradingAccount, err := r.TradingAccountRepository.Add(
				tx,
				req.UserID,
				institution,
				accountType,
				item.PlaidItemID,
				item.AccessToken,
				&acc.Name,
			)
			if err != nil {
				return fmt.Errorf("failed to add plaid account: %w", err)
			}

			mask := &acc.Mask
			if *mask == "" {
				mask = nil
			}
			err = r.TradingAccountRepository.AddPlaidMetadata(
				tx,
				tradingAccount.TradingAccountID,
				item.ItemID,
				mask,
			)
			if err != nil {
				return fmt.Errorf("failed to add plaid account metadata: %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func institutionFromPlaid(s string) model.CustodianType {
	switch s {
	case "Vanguard":
		return model.CustodianType_Vanguard
	}

	return model.CustodianType_Unknown
}

func accountTypeFromPlaid(s string) model.AccountType {
	switch s {
	case "401k":
		return model.AccountType_401k
	case "ira":
		return model.AccountType_Ira
	}

	return model.AccountType_Unknown
}
