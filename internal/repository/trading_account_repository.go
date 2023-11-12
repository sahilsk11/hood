package repository

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"time"

	"github.com/google/uuid"
)

type TradingAccountRepository interface {
	Add(tx *sql.Tx, userID uuid.UUID, custodian model.CustodianType, accountType model.AccountType, plaidItemID, accessToken string, name *string) (*model.TradingAccount, error)
	AddPlaidMetadata(tx *sql.Tx, tradingAccountID, itemID uuid.UUID, plaidAccountID string, mask *string) error
}

type tradingAccountRepositoryHandler struct {
	DB *sql.DB
}

func NewTradingAccountRepository(db *sql.DB) TradingAccountRepository {
	return tradingAccountRepositoryHandler{
		DB: db,
	}
}

func (h tradingAccountRepositoryHandler) Add(
	tx *sql.Tx,
	userID uuid.UUID,
	custodian model.CustodianType,
	accountType model.AccountType,
	plaidItemID,
	accessToken string,
	name *string,
) (*model.TradingAccount, error) {
	// TODO - update migration so we use generated uuid
	query := TradingAccount.INSERT(
		TradingAccount.AllColumns,
	).MODEL(
		model.TradingAccount{
			TradingAccountID: uuid.New(),
			UserID:           userID,
			Custodian:        custodian,
			AccountType:      accountType,
			Name:             name,
			CreatedAt:        time.Now().UTC(),
		},
	).RETURNING(
		TradingAccount.AllColumns,
	)

	tradingAccount := &model.TradingAccount{}
	err := query.Query(tx, tradingAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to insert trading account for %s: %w", userID.String(), err)
	}

	return tradingAccount, nil
}

func (h tradingAccountRepositoryHandler) AddPlaidMetadata(tx *sql.Tx, tradingAccountID, itemID uuid.UUID, plaidAccountID string, mask *string) error {
	query := PlaidTradingAccountMetadata.INSERT(
		PlaidTradingAccountMetadata.MutableColumns,
	).MODEL(
		model.PlaidTradingAccountMetadata{
			TradingAccountID: tradingAccountID,
			ItemID:           itemID,
			Mask:             mask,
			PlaidAccountID:   plaidAccountID,
		},
	)

	_, err := query.Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to insert plaid account metadata for %s: %w", tradingAccountID.String(), err)
	}

	return nil
}
