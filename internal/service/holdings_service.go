package service

import (
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	. "hood/internal/domain"
	"sort"

	"github.com/google/uuid"
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
		// Cash:     decimal.Zero,
	}.ToHoldings(), nil
}

func ConstructHistoricPortfolio(trades []model.Trade) (*domain.HistoricPortfolio, error) {
	// sanity check input
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].Date.After(trades[j].Date)
	})

	// getting so much deja vu with this
	// we've implemented this so many times.
	// tbh just find Playback. the only thing
	// im worried about is that i don't love
	// the mechanism for retrieving transfers

	// dude we need asset splits to do this
	// get Playback back

	return nil, nil
}
