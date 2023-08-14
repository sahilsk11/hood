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

func RollbackWithError(tx *sql.Tx, savepointName string, err error) error {
	if err != nil {
		if savepointErr := RollbackToSavepoint(savepointName, tx); savepointErr != nil {
			return fmt.Errorf("failed to rollback tx with err %w while handling error: %w", savepointErr, err)
		}
		return err
	}
	return nil
}
