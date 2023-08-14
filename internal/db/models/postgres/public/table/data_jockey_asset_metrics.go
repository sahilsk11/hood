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

var DataJockeyAssetMetrics = newDataJockeyAssetMetricsTable("public", "data_jockey_asset_metrics", "")

type dataJockeyAssetMetricsTable struct {
	postgres.Table

	//Columns
	ID        postgres.ColumnInteger
	JSON      postgres.ColumnString
	CreatedAt postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type DataJockeyAssetMetricsTable struct {
	dataJockeyAssetMetricsTable

	EXCLUDED dataJockeyAssetMetricsTable
}

// AS creates new DataJockeyAssetMetricsTable with assigned alias
func (a DataJockeyAssetMetricsTable) AS(alias string) *DataJockeyAssetMetricsTable {
	return newDataJockeyAssetMetricsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new DataJockeyAssetMetricsTable with assigned schema name
func (a DataJockeyAssetMetricsTable) FromSchema(schemaName string) *DataJockeyAssetMetricsTable {
	return newDataJockeyAssetMetricsTable(schemaName, a.TableName(), a.Alias())
}

func newDataJockeyAssetMetricsTable(schemaName, tableName, alias string) *DataJockeyAssetMetricsTable {
	return &DataJockeyAssetMetricsTable{
		dataJockeyAssetMetricsTable: newDataJockeyAssetMetricsTableImpl(schemaName, tableName, alias),
		EXCLUDED:                    newDataJockeyAssetMetricsTableImpl("", "excluded", ""),
	}
}

func newDataJockeyAssetMetricsTableImpl(schemaName, tableName, alias string) dataJockeyAssetMetricsTable {
	var (
		IDColumn        = postgres.IntegerColumn("id")
		JSONColumn      = postgres.StringColumn("json")
		CreatedAtColumn = postgres.TimestampzColumn("created_at")
		allColumns      = postgres.ColumnList{IDColumn, JSONColumn, CreatedAtColumn}
		mutableColumns  = postgres.ColumnList{JSONColumn, CreatedAtColumn}
	)

	return dataJockeyAssetMetricsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:        IDColumn,
		JSON:      JSONColumn,
		CreatedAt: CreatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
