package service

import (
	"context"
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	"hood/internal/domain"
	"hood/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// this might cause issues with MDS
// if asset is not available
const defaultTradeDate = "2010-01-01"

type IngestionService interface {
	AddPlaidTradeData(tx *sql.Tx, itemID uuid.UUID) error
}

func NewIngestionService(
	plaidRepository repository.PlaidRepository,
	tradeRepository repository.TradeRepository,
	plaidItemRepository repository.PlaidItemRepository,
	tradingAccountRepository repository.TradingAccountRepository,
	plaidInvestmentsAccountRepository repository.PlaidInvestmentsHoldingsRepository,
) IngestionService {
	return ingestionServiceHandler{
		PlaidRepository:                    plaidRepository,
		TradeRepository:                    tradeRepository,
		PlaidItemRepository:                plaidItemRepository,
		TradingAccountRepository:           tradingAccountRepository,
		PlaidInvestmentsHoldingsRepository: plaidInvestmentsAccountRepository,
	}
}

type ingestionServiceHandler struct {
	PlaidRepository                    repository.PlaidRepository
	TradeRepository                    repository.TradeRepository
	PlaidItemRepository                repository.PlaidItemRepository
	TradingAccountRepository           repository.TradingAccountRepository
	PlaidInvestmentsHoldingsRepository repository.PlaidInvestmentsHoldingsRepository
}

func (h ingestionServiceHandler) AddPlaidTradeData(tx *sql.Tx, itemID uuid.UUID) error {
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

	holdings, err := h.PlaidRepository.GetHoldings(
		item.AccessToken,
		plaidAccountIdToAccountID,
	)
	if err != nil {
		return err
	}

	err = h.PlaidInvestmentsHoldingsRepository.Add(tx, holdings)
	if err != nil {
		return err
	}

	defaultTradeDate, err := time.Parse(time.DateOnly, defaultTradeDate)
	if err != nil {
		return err
	}

	inferredTrades := []domain.Trade{}
	for _, tradingAccount := range plaidTradingAccounts {
		tradingAccountID := tradingAccount.TradingAccountID
		relevantTrades := []domain.Trade{}
		for _, t := range trades {
			if t.TradingAccountID == tradingAccount.TradingAccountID {
				relevantTrades = append(relevantTrades, t)
			}
		}
		startPortfolio := inverseTrades(relevantTrades, holdings[tradingAccountID])
		for _, p := range startPortfolio.Positions {
			inferredTrades = append(inferredTrades, domain.Trade{
				Symbol:           p.Symbol,
				Quantity:         p.Quantity,
				Price:            p.TotalCostBasis.Div(p.Quantity),
				Date:             defaultTradeDate, // day before their last trade, or beginning of time ?
				Description:      nil,
				TradingAccountID: tradingAccountID,
				Action:           model.TradeActionType_Buy,
				Source:           model.TradeSourceType_PlaidInferred,
			})
		}
	}

	_, err = h.TradeRepository.Add(tx, inferredTrades)
	if err != nil {
		return err
	}

	return nil
}

// inverseTrades takes a final portfolio holdings and a list of all trades
// we know the account did, then inverses each trade. the returned holdings
// is what the portfolio looks liked before all the given trades were made
//
// note - ensure the trades passed in actually belong to this holding
func inverseTrades(trades []domain.Trade, holdingsEnd domain.Holdings) domain.Holdings {
	// pretty sure order of trades doesn't matter
	// todo - handle cash

	out := &holdingsEnd
	for _, t := range trades {
		inv := decimal.NewFromInt(1)
		if t.Action == model.TradeActionType_Buy {
			inv = inv.Neg()
		}

		// when a trade sells to close, the position would disappear
		// from holdings. add it back if it's missing
		if _, ok := out.Positions[t.Symbol]; !ok {
			out.Positions[t.Symbol] = &domain.Position{
				Symbol: t.Symbol,
				// cost basis and quantity automatically set to 0
			}
		}

		// if action buy, inv = -1 and reduce stuff
		out.Positions[t.Symbol].Quantity = out.Positions[t.Symbol].Quantity.
			Add(inv.Mul(t.Quantity))
		tradeValue := t.Quantity.Mul(t.Price)
		out.Positions[t.Symbol].TotalCostBasis = out.Positions[t.Symbol].TotalCostBasis.
			Add(inv.Mul(tradeValue))

	}

	symbols := out.Symbols()

	for _, s := range symbols {
		if out.Positions[s].Quantity.IsZero() {
			delete(out.Positions, s)
		}
	}

	return *out
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
