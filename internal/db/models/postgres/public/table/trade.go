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

var Trade = newTradeTable("public", "trade", "")

type tradeTable struct {
	postgres.Table

	//Columns
	TradeID          postgres.ColumnInteger
	Symbol           postgres.ColumnString
	Action           postgres.ColumnString
	Quantity         postgres.ColumnFloat
	CostBasis        postgres.ColumnFloat
	Date             postgres.ColumnTimestampz
	Description      postgres.ColumnString
	CreatedAt        postgres.ColumnTimestampz
	ModifiedAt       postgres.ColumnTimestampz
	TradingAccountID postgres.ColumnString
	Source           postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type TradeTable struct {
	tradeTable

	EXCLUDED tradeTable
}

// AS creates new TradeTable with assigned alias
func (a TradeTable) AS(alias string) *TradeTable {
	return newTradeTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new TradeTable with assigned schema name
func (a TradeTable) FromSchema(schemaName string) *TradeTable {
	return newTradeTable(schemaName, a.TableName(), a.Alias())
}

func newTradeTable(schemaName, tableName, alias string) *TradeTable {
	return &TradeTable{
		tradeTable: newTradeTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newTradeTableImpl("", "excluded", ""),
	}
}

func newTradeTableImpl(schemaName, tableName, alias string) tradeTable {
	var (
		TradeIDColumn          = postgres.IntegerColumn("trade_id")
		SymbolColumn           = postgres.StringColumn("symbol")
		ActionColumn           = postgres.StringColumn("action")
		QuantityColumn         = postgres.FloatColumn("quantity")
		CostBasisColumn        = postgres.FloatColumn("cost_basis")
		DateColumn             = postgres.TimestampzColumn("date")
		DescriptionColumn      = postgres.StringColumn("description")
		CreatedAtColumn        = postgres.TimestampzColumn("created_at")
		ModifiedAtColumn       = postgres.TimestampzColumn("modified_at")
		TradingAccountIDColumn = postgres.StringColumn("trading_account_id")
		SourceColumn           = postgres.StringColumn("source")
		allColumns             = postgres.ColumnList{TradeIDColumn, SymbolColumn, ActionColumn, QuantityColumn, CostBasisColumn, DateColumn, DescriptionColumn, CreatedAtColumn, ModifiedAtColumn, TradingAccountIDColumn, SourceColumn}
		mutableColumns         = postgres.ColumnList{SymbolColumn, ActionColumn, QuantityColumn, CostBasisColumn, DateColumn, DescriptionColumn, CreatedAtColumn, ModifiedAtColumn, TradingAccountIDColumn, SourceColumn}
	)

	return tradeTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		TradeID:          TradeIDColumn,
		Symbol:           SymbolColumn,
		Action:           ActionColumn,
		Quantity:         QuantityColumn,
		CostBasis:        CostBasisColumn,
		Date:             DateColumn,
		Description:      DescriptionColumn,
		CreatedAt:        CreatedAtColumn,
		ModifiedAt:       ModifiedAtColumn,
		TradingAccountID: TradingAccountIDColumn,
		Source:           SourceColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
