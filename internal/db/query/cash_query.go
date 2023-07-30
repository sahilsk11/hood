package db

import (
	"database/sql"
	"fmt"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"time"
)

func AddCash(tx *sql.Tx, c model.Cash) error {
	c.CreatedAt = time.Now().UTC()
	query := Cash.INSERT(Cash.MutableColumns).
		MODEL(c)

	_, err := query.Exec(tx)
	if err != nil {
		return fmt.Errorf("failed to insert cash values into db: %w", err)
	}

	return nil
}
