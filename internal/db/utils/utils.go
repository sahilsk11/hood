package db_utils

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func GetTx(ctx context.Context) (*sql.Tx, error) {
	txVal := ctx.Value("tx")
	if txVal == nil {
		return nil, errors.New("no tx associated with request")
	}

	tx, ok := txVal.(*sql.Tx)
	if !ok {
		return nil, errors.New("could not cast context's tx to valid transaction")
	}

	return tx, nil
}

func IsDuplicateEntryErr(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
