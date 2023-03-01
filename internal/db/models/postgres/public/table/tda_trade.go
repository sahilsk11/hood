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

var TdaTrade = newTdaTradeTable("public", "tda_trade", "")

type tdaTradeTable struct {
	postgres.Table

	//Columns
	TdaTradeID       postgres.ColumnInteger
	TdaTransactionID postgres.ColumnInteger
	TradeID          postgres.ColumnInteger

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type TdaTradeTable struct {
	tdaTradeTable

	EXCLUDED tdaTradeTable
}

// AS creates new TdaTradeTable with assigned alias
func (a TdaTradeTable) AS(alias string) *TdaTradeTable {
	return newTdaTradeTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new TdaTradeTable with assigned schema name
func (a TdaTradeTable) FromSchema(schemaName string) *TdaTradeTable {
	return newTdaTradeTable(schemaName, a.TableName(), a.Alias())
}

func newTdaTradeTable(schemaName, tableName, alias string) *TdaTradeTable {
	return &TdaTradeTable{
		tdaTradeTable: newTdaTradeTableImpl(schemaName, tableName, alias),
		EXCLUDED:      newTdaTradeTableImpl("", "excluded", ""),
	}
}

func newTdaTradeTableImpl(schemaName, tableName, alias string) tdaTradeTable {
	var (
		TdaTradeIDColumn       = postgres.IntegerColumn("tda_trade_id")
		TdaTransactionIDColumn = postgres.IntegerColumn("tda_transaction_id")
		TradeIDColumn          = postgres.IntegerColumn("trade_id")
		allColumns             = postgres.ColumnList{TdaTradeIDColumn, TdaTransactionIDColumn, TradeIDColumn}
		mutableColumns         = postgres.ColumnList{TdaTransactionIDColumn, TradeIDColumn}
	)

	return tdaTradeTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		TdaTradeID:       TdaTradeIDColumn,
		TdaTransactionID: TdaTransactionIDColumn,
		TradeID:          TradeIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
