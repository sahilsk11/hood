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

var PlaidTradingAccountMetadata = newPlaidTradingAccountMetadataTable("public", "plaid_trading_account_metadata", "")

type plaidTradingAccountMetadataTable struct {
	postgres.Table

	//Columns
	PlaidAccountMetadataID postgres.ColumnString
	TradingAccountID       postgres.ColumnString
	Mask                   postgres.ColumnString
	ItemID                 postgres.ColumnString
	PlaidAccountID         postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type PlaidTradingAccountMetadataTable struct {
	plaidTradingAccountMetadataTable

	EXCLUDED plaidTradingAccountMetadataTable
}

// AS creates new PlaidTradingAccountMetadataTable with assigned alias
func (a PlaidTradingAccountMetadataTable) AS(alias string) *PlaidTradingAccountMetadataTable {
	return newPlaidTradingAccountMetadataTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new PlaidTradingAccountMetadataTable with assigned schema name
func (a PlaidTradingAccountMetadataTable) FromSchema(schemaName string) *PlaidTradingAccountMetadataTable {
	return newPlaidTradingAccountMetadataTable(schemaName, a.TableName(), a.Alias())
}

func newPlaidTradingAccountMetadataTable(schemaName, tableName, alias string) *PlaidTradingAccountMetadataTable {
	return &PlaidTradingAccountMetadataTable{
		plaidTradingAccountMetadataTable: newPlaidTradingAccountMetadataTableImpl(schemaName, tableName, alias),
		EXCLUDED:                         newPlaidTradingAccountMetadataTableImpl("", "excluded", ""),
	}
}

func newPlaidTradingAccountMetadataTableImpl(schemaName, tableName, alias string) plaidTradingAccountMetadataTable {
	var (
		PlaidAccountMetadataIDColumn = postgres.StringColumn("plaid_account_metadata_id")
		TradingAccountIDColumn       = postgres.StringColumn("trading_account_id")
		MaskColumn                   = postgres.StringColumn("mask")
		ItemIDColumn                 = postgres.StringColumn("item_id")
		PlaidAccountIDColumn         = postgres.StringColumn("plaid_account_id")
		allColumns                   = postgres.ColumnList{PlaidAccountMetadataIDColumn, TradingAccountIDColumn, MaskColumn, ItemIDColumn, PlaidAccountIDColumn}
		mutableColumns               = postgres.ColumnList{TradingAccountIDColumn, MaskColumn, ItemIDColumn, PlaidAccountIDColumn}
	)

	return plaidTradingAccountMetadataTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		PlaidAccountMetadataID: PlaidAccountMetadataIDColumn,
		TradingAccountID:       TradingAccountIDColumn,
		Mask:                   MaskColumn,
		ItemID:                 ItemIDColumn,
		PlaidAccountID:         PlaidAccountIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}