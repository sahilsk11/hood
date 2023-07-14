package hood_errors

import (
	"fmt"
	"hood/internal/db/models/postgres/public/model"
)

type ErrDuplicateTrade struct {
	Custodian              model.CustodianType
	CustodianTransactionID int64
	Message                string
}

func (e ErrDuplicateTrade) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("duplicate %s trade: %s", e.Custodian, e.Message)
	}
	return fmt.Sprintf("attempted to insert duplicate transaction of custodian %s with custodian ID %d", e.Custodian, e.CustodianTransactionID)
}
