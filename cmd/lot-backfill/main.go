package main

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"
	"hood/internal/portfolio"
	"log"
)

func getData(tx *sql.Tx, custodian model.CustodianType) portfolio.Events {
	trades, err := db.GetHistoricTrades(tx, custodian)
	if err != nil {
		log.Fatal(err)
	}
	assetSplits, err := db.GetHistoricAssetSplits(tx)
	if err != nil {
		log.Fatal(err)
	}
	transfers, err := db.GetHistoricTransfers(tx, custodian)
	if err != nil {
		log.Fatal(err)
	}
	dividends, err := db.GetHistoricDividends(tx, custodian)
	if err != nil {
		log.Fatal(err)
	}

	return portfolio.Events{
		Trades:      trades,
		AssetSplits: assetSplits,
		Transfers:   transfers,
		Dividends:   dividends,
	}
}

func main() {
	dbConn, err := db.New()
	e(err)
	tx, err := dbConn.Begin()
	e(err)

	events := getData(tx, model.CustodianType_Robinhood)
	dailyPortfolios, err := portfolio.PlaybackDaily(events)
	e(err)

	for dateStr, portfolio := range dailyPortfolios {
		err = insert(tx, portfolio)
		if err != nil {
			e(fmt.Errorf("failed insert portfolio on %s: %w", dateStr, err))
		}
	}

	err = tx.Commit()
	e(err)
}

func e(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func insert(tx *sql.Tx, portfolio domain.Portfolio) error {
	cash := portfolio.Cash
	err := db.AddCash(tx, model.Cash{
		Amount:    cash,
		Custodian: model.CustodianType_Robinhood,
		Date:      portfolio.LastAction,
	})
	if err != nil {
		return err
	}
	openLots := []domain.OpenLot{}
	for _, lots := range portfolio.OpenLots {
		for _, lot := range lots {
			openLots = append(openLots, *lot)
		}
	}
	for _, lot := range portfolio.NewOpenLots {
		openLots = append(openLots, lot)
	}
	err = db.AddImmutableOpenLots(tx, openLots)
	if err != nil {
		return err
	}
	return nil
}
