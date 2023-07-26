package db

import (
	"database/sql"
	"hood/internal/db/models/postgres/public/model"
	. "hood/internal/db/models/postgres/public/table"
	"hood/internal/domain"

	"github.com/go-jet/jet/v2/postgres"
)

func GetHistoricDividends(tx *sql.Tx, custodian model.CustodianType) ([]domain.Dividend, error) {
	query := Dividend.SELECT(Dividend.AllColumns).
		WHERE(
			Dividend.Custodian.EQ(postgres.NewEnumValue(custodian.String())),
		)

	result := []model.Dividend{}
	err := query.Query(tx, &result)
	if err != nil {
		return nil, err
	}
	return dividendsFromDb(result), nil
}

func dividendsFromDb(divs []model.Dividend) []domain.Dividend {
	out := make([]domain.Dividend, len(divs))
	for i, d := range divs {
		out[i] = domain.Dividend{
			Amount:              d.Amount,
			Date:                d.Date,
			Symbol:              d.Symbol,
			DividendID:          &d.DividendID,
			Custodian:           d.Custodian,
			ReinvestmentTradeID: d.ReinvestmentTradeID,
		}
	}
	return out
}
