//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package enum

import "github.com/go-jet/jet/v2/postgres"

var CustodianType = &struct {
	Robinhood postgres.StringExpression
	Tda       postgres.StringExpression
}{
	Robinhood: postgres.NewEnumValue("ROBINHOOD"),
	Tda:       postgres.NewEnumValue("TDA"),
}
