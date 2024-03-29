//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import "errors"

type TradeSourceType string

const (
	TradeSourceType_Manual        TradeSourceType = "MANUAL"
	TradeSourceType_Plaid         TradeSourceType = "PLAID"
	TradeSourceType_PlaidInferred TradeSourceType = "PLAID_INFERRED"
)

func (e *TradeSourceType) Scan(value interface{}) error {
	var enumValue string
	switch val := value.(type) {
	case string:
		enumValue = val
	case []byte:
		enumValue = string(val)
	default:
		return errors.New("jet: Invalid scan value for AllTypesEnum enum. Enum value has to be of type string or []byte")
	}

	switch enumValue {
	case "MANUAL":
		*e = TradeSourceType_Manual
	case "PLAID":
		*e = TradeSourceType_Plaid
	case "PLAID_INFERRED":
		*e = TradeSourceType_PlaidInferred
	default:
		return errors.New("jet: Invalid scan value '" + enumValue + "' for TradeSourceType enum")
	}

	return nil
}

func (e TradeSourceType) String() string {
	return string(e)
}
