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

var Price = newPriceTable("public", "price", "")

type priceTable struct {
	postgres.Table

	//Columns
	PriceID   postgres.ColumnInteger
	Symbol    postgres.ColumnString
	Price     postgres.ColumnFloat
	UpdatedAt postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type PriceTable struct {
	priceTable

	EXCLUDED priceTable
}

// AS creates new PriceTable with assigned alias
func (a PriceTable) AS(alias string) *PriceTable {
	return newPriceTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new PriceTable with assigned schema name
func (a PriceTable) FromSchema(schemaName string) *PriceTable {
	return newPriceTable(schemaName, a.TableName(), a.Alias())
}

func newPriceTable(schemaName, tableName, alias string) *PriceTable {
	return &PriceTable{
		priceTable: newPriceTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newPriceTableImpl("", "excluded", ""),
	}
}

func newPriceTableImpl(schemaName, tableName, alias string) priceTable {
	var (
		PriceIDColumn   = postgres.IntegerColumn("price_id")
		SymbolColumn    = postgres.StringColumn("symbol")
		PriceColumn     = postgres.FloatColumn("price")
		UpdatedAtColumn = postgres.TimestampzColumn("updated_at")
		allColumns      = postgres.ColumnList{PriceIDColumn, SymbolColumn, PriceColumn, UpdatedAtColumn}
		mutableColumns  = postgres.ColumnList{SymbolColumn, PriceColumn, UpdatedAtColumn}
	)

	return priceTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		PriceID:   PriceIDColumn,
		Symbol:    SymbolColumn,
		Price:     PriceColumn,
		UpdatedAt: UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
