package resolver

import (
	"context"

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

func (r Resolver) GeneratePlaidAccessToken(ctx context.Context, req api_types.GeneratePlaidAccessTokenRequest) error {
	accessToken, itemID, err := r.PlaidRepository.GetAccessToken(req.PublicToken)
	if err != nil {
		return err
	}

	_, err = r.PlaidItemRepository.Add(req.UserID, itemID, accessToken)
	if err != nil {
		return err
	}

	return nil
}
