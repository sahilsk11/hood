//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package enum

import "github.com/go-jet/jet/v2/postgres"

var BankActivityType = &struct {
	Withdrawal postgres.StringExpression
	Deposit    postgres.StringExpression
}{
	Withdrawal: postgres.NewEnumValue("WITHDRAWAL"),
	Deposit:    postgres.NewEnumValue("DEPOSIT"),
}
