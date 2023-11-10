package db

import (
	"database/sql"
	. "hood/internal/db/models/postgres/public/table"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/shopspring/decimal"
)

func GetTotalRealizedCostBasis(tx *sql.Tx) (decimal.Decimal, error) {
	// calculate the total (purchase price * quantity) for all realized assets
	// purchase price = sell price - realized_gains/closed_lot.quantity
	// this is because buy.purchase_price may fluctuate with stock splits

	// TODO: consider querying everything into code and using decimal for math
	purchasePriceExp := Trade.CostBasis.SUB(ClosedLot.RealizedGains.DIV(ClosedLot.Quantity))
	query := ClosedLot.SELECT(SUM(
		purchasePriceExp.MUL(ClosedLot.Quantity), // total cost basis for single closed lot
	)).
		WHERE(ClosedLot.Quantity.GT(Float(0))).
		FROM(
			ClosedLot.INNER_JOIN(Trade, ClosedLot.SellTradeID.EQ(Trade.TradeID)),
		)

	return fetchDecimal(tx, query)
}

func GetTotalRealizedGains(tx *sql.Tx) (decimal.Decimal, error) {
	return fetchDecimal(tx, ClosedLot.SELECT(SUM(ClosedLot.RealizedGains)))
}

func GetTotalUnrealizedCostBasis(tx *sql.Tx) (decimal.Decimal, error) {
	return fetchDecimal(tx, OpenLot.SELECT(SUM(OpenLot.CostBasis.MUL(OpenLot.Quantity))))
}

func fetchDecimal(tx *sql.Tx, q Statement) (decimal.Decimal, error) {
	query, args := q.Sql()
	row := tx.QueryRow(query, args...)

	var total float64
	err := row.Scan(&total)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromFloat(total), nil
}
