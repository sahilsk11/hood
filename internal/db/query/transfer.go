package db

import (
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
)

func GetHistoricTransfers(tx *sql.Tx) ([]model.Cash, error) {
	query := Cash.SELECT(Cash.AllColumns).ORDER_BY(Cash.Date.ASC())
	out := []model.Cash{}
	err := query.Query(tx, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func AddTransfer(tx *sql.Tx, t *model.Cash) error {
	query := Cash.INSERT(Cash.MutableColumns).MODEL(t).RETURNING(Cash.AllColumns)
	err := query.Query(tx, t)
	if err != nil {
		return err
	}
	return nil
}
