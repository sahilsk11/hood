package service

import (
	"database/sql"
	db "hood/internal/db/query"
	. "hood/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// current cash calculation is broken
const rhCashOverride = 130
const tdaCashOverride = 5000

var cashOverrides = map[string]float64{}

func GetCurrentPortfolio(tx *sql.Tx, tradingAccountID uuid.UUID) (*Portfolio, error) {
	openLots, err := db.GetCurrentOpenLots(tx, tradingAccountID)
	if err != nil {
		return nil, err
	}

	// TODO - fix
	cashOverride := decimal.Zero
	if override, ok := cashOverrides[tradingAccountID.String()]; ok {
		cashOverride = decimal.NewFromFloat(override)
	}

	mappedOpenLots := map[string][]*OpenLot{}
	for _, lot := range openLots {
		symbol := lot.GetSymbol()
		if _, ok := mappedOpenLots[symbol]; !ok {
			mappedOpenLots[symbol] = []*OpenLot{}
		}
		mappedOpenLots[symbol] = append(mappedOpenLots[symbol], &lot)
	}

	return &Portfolio{
		OpenLots: mappedOpenLots,
		Cash:     cashOverride,
		// LastAction: , TODO - how to populate this field
	}, nil
}

func GetAggregatePortfolio(tx *sql.Tx, userID uuid.UUID) (*Portfolio, error) {
	// add queries

	return nil, nil
}
