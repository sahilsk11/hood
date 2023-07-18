package db

import (
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
)

func GetHistoricTransfers(tx *sql.Tx) ([]model.BankActivity, error) {
	query := BankActivity.SELECT(BankActivity.AllColumns)
	out := []model.BankActivity{}
	err := query.Query(tx, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func AddTransfer(tx *sql.Tx, t *model.BankActivity) error {
	query := BankActivity.INSERT(BankActivity.MutableColumns).MODEL(t).RETURNING(BankActivity.AllColumns)
	err := query.Query(tx, t)
	if err != nil {
		return err
	}
	return nil
}
