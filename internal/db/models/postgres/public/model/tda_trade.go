//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

type TdaTrade struct {
	TdaTradeID       int32 `sql:"primary_key"`
	TdaTransactionID int64
	TradeID          int32
}
