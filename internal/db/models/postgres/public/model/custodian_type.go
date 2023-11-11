//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import "errors"

type CustodianType string

const (
	CustodianType_Robinhood     CustodianType = "ROBINHOOD"
	CustodianType_Tda           CustodianType = "TDA"
	CustodianType_Schwab        CustodianType = "SCHWAB"
	CustodianType_Vanguard      CustodianType = "VANGUARD"
	CustodianType_Wealthfront   CustodianType = "WEALTHFRONT"
	CustodianType_Betterment    CustodianType = "BETTERMENT"
	CustodianType_MorganStanley CustodianType = "MORGAN_STANLEY"
	CustodianType_ETrade        CustodianType = "E-TRADE"
	CustodianType_Unknown       CustodianType = "UNKNOWN"
)

func (e *CustodianType) Scan(value interface{}) error {
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
	case "ROBINHOOD":
		*e = CustodianType_Robinhood
	case "TDA":
		*e = CustodianType_Tda
	case "SCHWAB":
		*e = CustodianType_Schwab
	case "VANGUARD":
		*e = CustodianType_Vanguard
	case "WEALTHFRONT":
		*e = CustodianType_Wealthfront
	case "BETTERMENT":
		*e = CustodianType_Betterment
	case "MORGAN_STANLEY":
		*e = CustodianType_MorganStanley
	case "E-TRADE":
		*e = CustodianType_ETrade
	case "UNKNOWN":
		*e = CustodianType_Unknown
	default:
		return errors.New("jet: Invalid scan value '" + enumValue + "' for CustodianType enum")
	}

	return nil
}

func (e CustodianType) String() string {
	return string(e)
}
