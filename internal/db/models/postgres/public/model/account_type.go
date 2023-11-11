//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import "errors"

type AccountType string

const (
	AccountType_Individual AccountType = "INDIVIDUAL"
	AccountType_Ira        AccountType = "IRA"
	AccountType_RothIra    AccountType = "ROTH_IRA"
	AccountType_401k       AccountType = "401k"
	AccountType_Unknown    AccountType = "UNKNOWN"
)

func (e *AccountType) Scan(value interface{}) error {
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
	case "INDIVIDUAL":
		*e = AccountType_Individual
	case "IRA":
		*e = AccountType_Ira
	case "ROTH_IRA":
		*e = AccountType_RothIra
	case "401k":
		*e = AccountType_401k
	case "UNKNOWN":
		*e = AccountType_Unknown
	default:
		return errors.New("jet: Invalid scan value '" + enumValue + "' for AccountType enum")
	}

	return nil
}

func (e AccountType) String() string {
	return string(e)
}
