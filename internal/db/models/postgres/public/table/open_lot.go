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

var OpenLot = newOpenLotTable("public", "open_lot", "")

type openLotTable struct {
	postgres.Table

	//Columns
	OpenLotID  postgres.ColumnInteger
	CostBasis  postgres.ColumnFloat
	Quantity   postgres.ColumnFloat
	TradeID    postgres.ColumnInteger
	DeletedAt  postgres.ColumnTimestampz
	CreatedAt  postgres.ColumnTimestampz
	ModifiedAt postgres.ColumnTimestampz
	LotID      postgres.ColumnString
	Date       postgres.ColumnDate

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type OpenLotTable struct {
	openLotTable

	EXCLUDED openLotTable
}

// AS creates new OpenLotTable with assigned alias
func (a OpenLotTable) AS(alias string) *OpenLotTable {
	return newOpenLotTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new OpenLotTable with assigned schema name
func (a OpenLotTable) FromSchema(schemaName string) *OpenLotTable {
	return newOpenLotTable(schemaName, a.TableName(), a.Alias())
}

func newOpenLotTable(schemaName, tableName, alias string) *OpenLotTable {
	return &OpenLotTable{
		openLotTable: newOpenLotTableImpl(schemaName, tableName, alias),
		EXCLUDED:     newOpenLotTableImpl("", "excluded", ""),
	}
}

func newOpenLotTableImpl(schemaName, tableName, alias string) openLotTable {
	var (
		OpenLotIDColumn  = postgres.IntegerColumn("open_lot_id")
		CostBasisColumn  = postgres.FloatColumn("cost_basis")
		QuantityColumn   = postgres.FloatColumn("quantity")
		TradeIDColumn    = postgres.IntegerColumn("trade_id")
		DeletedAtColumn  = postgres.TimestampzColumn("deleted_at")
		CreatedAtColumn  = postgres.TimestampzColumn("created_at")
		ModifiedAtColumn = postgres.TimestampzColumn("modified_at")
		LotIDColumn      = postgres.StringColumn("lot_id")
		DateColumn       = postgres.DateColumn("date")
		allColumns       = postgres.ColumnList{OpenLotIDColumn, CostBasisColumn, QuantityColumn, TradeIDColumn, DeletedAtColumn, CreatedAtColumn, ModifiedAtColumn, LotIDColumn, DateColumn}
		mutableColumns   = postgres.ColumnList{CostBasisColumn, QuantityColumn, TradeIDColumn, DeletedAtColumn, CreatedAtColumn, ModifiedAtColumn, LotIDColumn, DateColumn}
	)

	return openLotTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		OpenLotID:  OpenLotIDColumn,
		CostBasis:  CostBasisColumn,
		Quantity:   QuantityColumn,
		TradeID:    TradeIDColumn,
		DeletedAt:  DeletedAtColumn,
		CreatedAt:  CreatedAtColumn,
		ModifiedAt: ModifiedAtColumn,
		LotID:      LotIDColumn,
		Date:       DateColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
