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

var Position = newPositionTable("public", "position", "")

type positionTable struct {
	postgres.Table

	//Columns
	PositionID       postgres.ColumnString
	Ticker           postgres.ColumnString
	TradingAccountID postgres.ColumnString
	TotalCostBasis   postgres.ColumnFloat
	Quantity         postgres.ColumnFloat
	CreatedAt        postgres.ColumnTimestampz
	Source           postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type PositionTable struct {
	positionTable

	EXCLUDED positionTable
}

// AS creates new PositionTable with assigned alias
func (a PositionTable) AS(alias string) *PositionTable {
	return newPositionTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new PositionTable with assigned schema name
func (a PositionTable) FromSchema(schemaName string) *PositionTable {
	return newPositionTable(schemaName, a.TableName(), a.Alias())
}

func newPositionTable(schemaName, tableName, alias string) *PositionTable {
	return &PositionTable{
		positionTable: newPositionTableImpl(schemaName, tableName, alias),
		EXCLUDED:      newPositionTableImpl("", "excluded", ""),
	}
}

func newPositionTableImpl(schemaName, tableName, alias string) positionTable {
	var (
		PositionIDColumn       = postgres.StringColumn("position_id")
		TickerColumn           = postgres.StringColumn("ticker")
		TradingAccountIDColumn = postgres.StringColumn("trading_account_id")
		TotalCostBasisColumn   = postgres.FloatColumn("total_cost_basis")
		QuantityColumn         = postgres.FloatColumn("quantity")
		CreatedAtColumn        = postgres.TimestampzColumn("created_at")
		SourceColumn           = postgres.StringColumn("source")
		allColumns             = postgres.ColumnList{PositionIDColumn, TickerColumn, TradingAccountIDColumn, TotalCostBasisColumn, QuantityColumn, CreatedAtColumn, SourceColumn}
		mutableColumns         = postgres.ColumnList{TickerColumn, TradingAccountIDColumn, TotalCostBasisColumn, QuantityColumn, CreatedAtColumn, SourceColumn}
	)

	return positionTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		PositionID:       PositionIDColumn,
		Ticker:           TickerColumn,
		TradingAccountID: TradingAccountIDColumn,
		TotalCostBasis:   TotalCostBasisColumn,
		Quantity:         QuantityColumn,
		CreatedAt:        CreatedAtColumn,
		Source:           SourceColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
