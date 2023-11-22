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

// IngestionService is concerned with the addition and mutation of new portfolio data
type IngestionService interface {
	// finds all available trades for the account, adds them to db,
	// and guesses any missing trade data to match latest Plaid holdings
	AddPlaidTradeData(tx *sql.Tx, tradingAccountID uuid.UUID) error
	UpdatePosition(tx *sql.Tx, tradingAccountID uuid.UUID, newPosition domain.Position) error
}

func NewIngestionService(
	plaidRepository repository.PlaidRepository,
	tradeRepository repository.TradeRepository,
	plaidItemRepository repository.PlaidItemRepository,
	tradingAccountRepository repository.TradingAccountRepository,
	positionsRepository repository.PositionsRepository,
) IngestionService {
	return ingestionServiceHandler{
		PlaidRepository:          plaidRepository,
		TradeRepository:          tradeRepository,
		PlaidItemRepository:      plaidItemRepository,
		TradingAccountRepository: tradingAccountRepository,
		PositionsRepository:      positionsRepository,
	}
}

type ingestionServiceHandler struct {
	PlaidRepository          repository.PlaidRepository
	TradeRepository          repository.TradeRepository
	PlaidItemRepository      repository.PlaidItemRepository
	TradingAccountRepository repository.TradingAccountRepository
	PositionsRepository      repository.PositionsRepository
}

func (h ingestionServiceHandler) AddPlaidTradeData(tx *sql.Tx, tradingAccountID uuid.UUID) error {
	tradingAccount, err := h.TradingAccountRepository.Get(tx, tradingAccountID)
	if err != nil {
		return err
	}
	if tradingAccount.DataSource != model.TradingAccountDataSourceType_Trades {
		return fmt.Errorf("cannot add trades to account %s - data source is not trades", tradingAccountID.String())
	}

	// todo - we need some way to check if an account is actively connected via Plaid
	// knowing there's a Plaid metadata link is insufficient, because the link can
	// exist but not be active

	plaidTradingAccountMetadata, err := h.TradingAccountRepository.GetPlaidMetadata(tx, tradingAccountID)
	if err != nil {
		return err
	}
	if plaidTradingAccountMetadata == nil {
		return fmt.Errorf("could not find plaid metadata for trading account %s", tradingAccountID.String())
	}

	item, err := h.PlaidItemRepository.Get(tx, plaidTradingAccountMetadata.ItemID)
	if err != nil {
		return err
	}

	trades, plaidTrades, err := h.PlaidRepository.GetTransactions(
		context.Background(),
		// todo - make this param plaidTradingAccountMetadata
		map[string]uuid.UUID{
			plaidTradingAccountMetadata.PlaidAccountID: plaidTradingAccountMetadata.TradingAccountID,
		},
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
		// todo - make this param plaidTradingAccountMetadata
		map[string]uuid.UUID{
			plaidTradingAccountMetadata.PlaidAccountID: plaidTradingAccountMetadata.TradingAccountID,
		},
	)
	if err != nil {
		return err
	}

	err = h.PositionsRepository.Add(tx, holdings, model.PositionSourceType_Plaid)
	if err != nil {
		return err
	}

	defaultTradeDate, err := time.Parse(time.DateOnly, defaultTradeDate)
	if err != nil {
		return err
	}

	// this is wrong - we should pull trades we have for this account,
	// not just the ones we got from Plaid
	// TODO - i just don't know how to reconcile existing trades

	inferredTrades := []domain.Trade{}
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

func (h ingestionServiceHandler) UpdatePosition(tx *sql.Tx, tradingAccountID uuid.UUID, newPosition domain.Position) error {
	tradingAccount, err := h.TradingAccountRepository.Get(tx, tradingAccountID)
	if err != nil {
		return err
	}
	if tradingAccount.DataSource != model.TradingAccountDataSourceType_Positions {
		return fmt.Errorf("cannot update positions of %s - account data source is not positions", tradingAccountID.String())
	}

	positions, err := h.PositionsRepository.List(tx, tradingAccountID)
	if err != nil {
		return err
	}

	for _, p := range positions {
		if p.Symbol == newPosition.Symbol && !newPosition.Quantity.Sub(p.Quantity).IsZero() {
			err = h.PositionsRepository.Delete(tx, tradingAccountID, newPosition.Symbol)
			if err != nil {
				return err
			}
		}
	}

	// you actually gotta kys for this bs
	// TODO - fix
	err = h.PositionsRepository.Add(tx, map[uuid.UUID]domain.Holdings{
		tradingAccountID: {
			Positions: map[string]*domain.Position{
				newPosition.Symbol: &newPosition,
			},
		},
	},
		model.PositionSourceType_Manual,
	)
	if err != nil {
		return err
	}

	return nil
}
