package service

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	"hood/internal/repository"

	"github.com/google/uuid"
)

type IngestionService interface {
	AddPlaidTrades()
}

type ingestionServiceHandler struct {
	PlaidRepository          repository.PlaidRepository
	TradeRepository          repository.TradeRepository
	PlaidItemRepository      repository.PlaidItemRepository
	TradingAccountRepository repository.TradingAccountRepository
}

func (h ingestionServiceHandler) AddPlaidData(tx *sql.Tx, itemID uuid.UUID) error {
	item, err := h.PlaidItemRepository.Get(tx, itemID)
	if err != nil {
		return err
	}

	plaidTradingAccounts, err := h.TradingAccountRepository.GetPlaidMetadata(tx, itemID)
	if err != nil {
		return err
	}

	plaidAccountIdToAccountID := map[string]uuid.UUID{}
	for _, acc := range plaidTradingAccounts {
		plaidAccountIdToAccountID[acc.PlaidAccountID] = acc.TradingAccountID
	}

	trades, plaidTrades, err := h.PlaidRepository.GetTransactions(
		context.Background(),
		plaidAccountIdToAccountID,
		item.AccessToken,
	)
	if err != nil {
		return fmt.Errorf("failed to get plaid transactions: %w", err)
	}

	// adding everything to DB ensures we remove duplicates
	err = h.AddPlaidTrades(tx, trades, plaidTrades)
	if err != nil {
		return fmt.Errorf("failed to add plaid trades: %w", err)
	}

	return nil
}

// AddPlaidTrades adds trades that were retrieved from Plaid
// into the database. It's assumed that plaidTrades does not have
// tradeID populated since the og trades need to be added to the database
// first.
// TODO - consider how to reconcile with existing trades
func (h ingestionServiceHandler) AddPlaidTrades(
	tx *sql.Tx,
	trades []domain.Trade,
	plaidTrades []model.PlaidTradeMetadata,
) error {
	insertedTrades, err := h.TradeRepository.Add(
		tx,
		trades,
	)
	if err != nil {
		return err
	}

	for i, t := range insertedTrades {
		plaidTrades[i].TradeID = t.TradeID
	}

	err = h.TradeRepository.AddPlaidMetadata(tx, plaidTrades)
	if err != nil {
		return err
	}

	return nil
}
