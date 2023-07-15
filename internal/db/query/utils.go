package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
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

func New() (*sql.DB, error) {
	connStr := "postgresql://postgres:postgres@localhost:5438/postgres?sslmode=disable"
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

func NewTest() (*sql.DB, error) {
	connStr := "postgresql://postgres:postgres@localhost:5438/postgres_test?sslmode=disable"
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

func CleanupTest(t *testing.T, tx *sql.Tx) {
	t.Cleanup(func() {
		err := tx.Rollback()
		if err != nil {
			panic(err)
		}
	})
}
