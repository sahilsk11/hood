package service

import (
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	. "hood/internal/domain"

	"github.com/shopspring/decimal"
)

// current cash calculation is broken
const rhCashOverride = 130
const tdaCashOverride = 5000

func GetCurrentPortfolio(tx *sql.Tx, custodian model.CustodianType) (*Portfolio, error) {
	openLots, err := db.GetCurrentOpenLots(tx, custodian)
	if err != nil {
		return nil, err
	}

	var cash decimal.Decimal
	if custodian == model.CustodianType_Tda {
		cash = decimal.NewFromFloat(rhCashOverride)
	} else {
		cash = decimal.NewFromFloat(tdaCashOverride)
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
		Cash:     cash,
		// LastAction: , TODO - how to populate this field
	}, nil
}

func GetAggregatePortfolio(tx *sql.Tx) (*Portfolio, error) {
	tdaPortfolio, err := GetCurrentPortfolio(tx, model.CustodianType_Tda)
	if err != nil {
		return nil, err
	}
	rhPortfolio, err := GetCurrentPortfolio(tx, model.CustodianType_Robinhood)
	if err != nil {
		return nil, err
	}

	combinedPortfolio := tdaPortfolio.Add(*rhPortfolio)
	return &combinedPortfolio, nil
}
