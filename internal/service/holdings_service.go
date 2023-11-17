package service

import (
	"database/sql"
	db "hood/internal/db/query"
	. "hood/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type HoldingsService interface {
	Get(tx *sql.Tx, tradingAccountID uuid.UUID) (*Holdings, error)
}

type holdingsServiceHandler struct{}

func NewHoldingsService() HoldingsService {
	return holdingsServiceHandler{}
}

func (h holdingsServiceHandler) Get(tx *sql.Tx, tradingAccountID uuid.UUID) (*Holdings, error) {
	// fuckkkkk the open lot logic is broken
	// so i dont think this works. iirc the new
	// plaid ingestion code doesn't even try
	// to add open lots
	// .. maybe just nuke the tables and retry
	// this db structure is a huge pain. i think
	// we should tear it all away and re-run
	// all trades every time we want something
	openLots, err := db.GetCurrentOpenLots(tx, tradingAccountID)
	if err != nil {
		return nil, err
	}

	mappedOpenLots := map[string][]*OpenLot{}
	for _, lot := range openLots {
		symbol := lot.GetSymbol()
		if _, ok := mappedOpenLots[symbol]; !ok {
			mappedOpenLots[symbol] = []*OpenLot{}
		}
		mappedOpenLots[symbol] = append(mappedOpenLots[symbol], &lot)
	}

	return Portfolio{
		OpenLots: mappedOpenLots,
		Cash:     decimal.Zero,
	}.ToHoldings(), nil
}
