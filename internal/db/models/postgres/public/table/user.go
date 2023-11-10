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

var User = newUserTable("public", "user", "")

type userTable struct {
	postgres.Table

	//Columns
	UserID       postgres.ColumnString
	FirstName    postgres.ColumnString
	MiddleName   postgres.ColumnString
	LastName     postgres.ColumnString
	PrimaryEmail postgres.ColumnString
	CreatedAt    postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type UserTable struct {
	userTable

	EXCLUDED userTable
}

// AS creates new UserTable with assigned alias
func (a UserTable) AS(alias string) *UserTable {
	return newUserTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new UserTable with assigned schema name
func (a UserTable) FromSchema(schemaName string) *UserTable {
	return newUserTable(schemaName, a.TableName(), a.Alias())
}

func newUserTable(schemaName, tableName, alias string) *UserTable {
	return &UserTable{
		userTable: newUserTableImpl(schemaName, tableName, alias),
		EXCLUDED:  newUserTableImpl("", "excluded", ""),
	}
}

func newUserTableImpl(schemaName, tableName, alias string) userTable {
	var (
		UserIDColumn       = postgres.StringColumn("user_id")
		FirstNameColumn    = postgres.StringColumn("first_name")
		MiddleNameColumn   = postgres.StringColumn("middle_name")
		LastNameColumn     = postgres.StringColumn("last_name")
		PrimaryEmailColumn = postgres.StringColumn("primary_email")
		CreatedAtColumn    = postgres.TimestampzColumn("created_at")
		allColumns         = postgres.ColumnList{UserIDColumn, FirstNameColumn, MiddleNameColumn, LastNameColumn, PrimaryEmailColumn, CreatedAtColumn}
		mutableColumns     = postgres.ColumnList{FirstNameColumn, MiddleNameColumn, LastNameColumn, PrimaryEmailColumn, CreatedAtColumn}
	)

	return userTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		UserID:       UserIDColumn,
		FirstName:    FirstNameColumn,
		MiddleName:   MiddleNameColumn,
		LastName:     LastNameColumn,
		PrimaryEmail: PrimaryEmailColumn,
		CreatedAt:    CreatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
