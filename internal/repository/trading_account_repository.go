package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type TradingAccountRepository interface {
	Add(tx *sql.Tx, userID uuid.UUID, custodian model.CustodianType, accountType model.AccountType, name *string, dataSource model.TradingAccountDataSourceType) (*model.TradingAccount, error)
	Get(tx *sql.Tx, tradingAccountID uuid.UUID) (*model.TradingAccount, error)
	AddPlaidMetadata(tx *sql.Tx, tradingAccountID, itemID uuid.UUID, plaidAccountID string, mask *string) error
	GetPlaidMetadata(tx *sql.Tx, tradingAccountID uuid.UUID) (*model.PlaidTradingAccountMetadata, error)
	ListPlaidMetadataByItemID(tx *sql.Tx, itemID uuid.UUID) ([]model.PlaidTradingAccountMetadata, error)
}

type tradingAccountRepositoryHandler struct {
	DB *sql.DB
}

func NewTradingAccountRepository(db *sql.DB) TradingAccountRepository {
	return tradingAccountRepositoryHandler{
		DB: db,
	}
}

func (h tradingAccountRepositoryHandler) GetPlaidMetadata(tx *sql.Tx, tradingAccountID uuid.UUID) (*model.PlaidTradingAccountMetadata, error) {
	query := PlaidTradingAccountMetadata.SELECT(
		PlaidTradingAccountMetadata.AllColumns,
	).WHERE(
		PlaidTradingAccountMetadata.TradingAccountID.EQ(
			postgres.UUID(tradingAccountID),
		),
	)

	var out model.PlaidTradingAccountMetadata
	err := query.Query(tx, &out)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &out, nil
}

func (h tradingAccountRepositoryHandler) Get(tx *sql.Tx, tradingAccountID uuid.UUID) (*model.TradingAccount, error) {
	query := TradingAccount.SELECT(TradingAccount.AllColumns).
		WHERE(TradingAccount.TradingAccountID.EQ(postgres.UUID(tradingAccountID)))

	var out model.TradingAccount
	err := query.Query(tx, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (h tradingAccountRepositoryHandler) Add(
	tx *sql.Tx,
	userID uuid.UUID,
	custodian model.CustodianType,
	accountType model.AccountType,
	name *string,
	dataSource model.TradingAccountDataSourceType,
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
			DataSource:       dataSource,
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

func (h tradingAccountRepositoryHandler) ListPlaidMetadataByItemID(tx *sql.Tx, itemID uuid.UUID) ([]model.PlaidTradingAccountMetadata, error) {
	query := PlaidTradingAccountMetadata.SELECT(
		PlaidTradingAccountMetadata.AllColumns,
	).WHERE(
		PlaidTradingAccountMetadata.ItemID.EQ(postgres.UUID(itemID)),
	)

	models := []model.PlaidTradingAccountMetadata{}
	err := query.Query(tx, &models)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaid trading account metadata: %w", err)
	}

	return models, nil
}
