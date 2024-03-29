//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Dividend = newDividendTable("public", "dividend", "")

type dividendTable struct {
	postgres.Table

	//Columns
	DividendID          postgres.ColumnInteger
	Symbol              postgres.ColumnString
	Amount              postgres.ColumnFloat
	Date                postgres.ColumnDate
	Custodian           postgres.ColumnString
	ReinvestmentTradeID postgres.ColumnInteger
	CreatedAt           postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type DividendTable struct {
	dividendTable

	EXCLUDED dividendTable
}

// AS creates new DividendTable with assigned alias
func (a DividendTable) AS(alias string) *DividendTable {
	return newDividendTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new DividendTable with assigned schema name
func (a DividendTable) FromSchema(schemaName string) *DividendTable {
	return newDividendTable(schemaName, a.TableName(), a.Alias())
}

func newDividendTable(schemaName, tableName, alias string) *DividendTable {
	return &DividendTable{
		dividendTable: newDividendTableImpl(schemaName, tableName, alias),
		EXCLUDED:      newDividendTableImpl("", "excluded", ""),
	}
}

func newDividendTableImpl(schemaName, tableName, alias string) dividendTable {
	var (
		DividendIDColumn          = postgres.IntegerColumn("dividend_id")
		SymbolColumn              = postgres.StringColumn("symbol")
		AmountColumn              = postgres.FloatColumn("amount")
		DateColumn                = postgres.DateColumn("date")
		CustodianColumn           = postgres.StringColumn("custodian")
		ReinvestmentTradeIDColumn = postgres.IntegerColumn("reinvestment_trade_id")
		CreatedAtColumn           = postgres.TimestampzColumn("created_at")
		allColumns                = postgres.ColumnList{DividendIDColumn, SymbolColumn, AmountColumn, DateColumn, CustodianColumn, ReinvestmentTradeIDColumn, CreatedAtColumn}
		mutableColumns            = postgres.ColumnList{SymbolColumn, AmountColumn, DateColumn, CustodianColumn, ReinvestmentTradeIDColumn, CreatedAtColumn}
	)

	return dividendTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		DividendID:          DividendIDColumn,
		Symbol:              SymbolColumn,
		Amount:              AmountColumn,
		Date:                DateColumn,
		Custodian:           CustodianColumn,
		ReinvestmentTradeID: ReinvestmentTradeIDColumn,
		CreatedAt:           CreatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
