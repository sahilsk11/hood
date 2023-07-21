package db

import (
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
)

func GetHistoricTransfers(tx *sql.Tx) ([]model.Transfer, error) {
	query := Transfer.SELECT(Transfer.AllColumns).ORDER_BY(Transfer.Date.ASC())
	out := []model.Transfer{}
	err := query.Query(tx, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func AddTransfer(tx *sql.Tx, t *model.Transfer) error {
	query := Transfer.INSERT(Transfer.MutableColumns).MODEL(t).RETURNING(Transfer.AllColumns)
	err := query.Query(tx, t)
	if err != nil {
		return err
	}
	return nil
}
