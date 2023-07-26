package db

import (
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/domain"
)

func GetHistoricTransfers(tx *sql.Tx) ([]domain.Transfer, error) {
	query := Transfer.SELECT(Transfer.AllColumns).ORDER_BY(Transfer.Date.ASC())
	out := []model.Transfer{}
	err := query.Query(tx, &out)
	if err != nil {
		return nil, err
	}
	return transfersFromDb(out), nil
}

func transfersFromDb(transfers []model.Transfer) []domain.Transfer {
	out := make([]domain.Transfer, len(transfers))
	for i, t := range transfers {
		out[i] = domain.Transfer{
			ActivityID:   &t.ActivityID,
			Amount:       t.Amount,
			ActivityType: t.ActivityType,
			Date:         t.Date,
			Custodian:    t.Custodian,
		}
	}
	return out
}

func AddTransfer(tx *sql.Tx, t *model.Transfer) error {
	query := Transfer.INSERT(Transfer.MutableColumns).MODEL(t).RETURNING(Transfer.AllColumns)
	err := query.Query(tx, t)
	if err != nil {
		return err
	}
	return nil
}
