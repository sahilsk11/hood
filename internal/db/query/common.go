package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func AddSavepoint(tx *sql.Tx) (string, error) {
	savepointName := "x" + strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err := tx.Exec("SAVEPOINT " + savepointName + ";")
	if err != nil {
		return "", fmt.Errorf("failed to create savepoint: %w", err)
	}

	return savepointName, nil
}

func RollbackToSavepoint(name string, tx *sql.Tx) error {
	_, err := tx.Exec("ROLLBACK TO SAVEPOINT " + name)
	return err
}
