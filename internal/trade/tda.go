package trade

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	hood_errors "hood/internal"
	"hood/internal/db/models/postgres/public/model"
	db "hood/internal/db/query"
	"hood/internal/domain"

	"os"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func determineColumnOrder(headerRow []string) (map[string]int, error) {
	requiredColumns := []string{
		"date",
		"action",
		"quantity",
		"symbol",
		"price",
	}

	columnIndices := map[string]int{}
	for i, h := range headerRow {
		h = strings.ToLower(h)
		h = strings.ReplaceAll(h, " ", "_")
		for _, rc := range requiredColumns {
			if h == rc {
				columnIndices[h] = i
			}
		}
	}

	for _, rc := range requiredColumns {
		if _, ok := columnIndices[rc]; !ok {
			return nil, fmt.Errorf("missing required column '%s'", rc)
		}
	}

	return columnIndices, nil
}

func ParseSchwabTransactionFile(ctx context.Context, tx *sql.Tx, csvFileName string, tiService TradeIngestionService) ([]domain.Trade, error) {
	f, err := os.Open(csvFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvFile := csv.NewReader(f)
	records, err := csvFile.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv file: %w", err)
	}

	orders := []domain.Trade{}

	ordering, err := determineColumnOrder(records[0])
	if err != nil {
		return nil, err
	}

	for _, record := range records[1:] {
		actionStr := strings.ToLower(record[ordering["action"]])
		if strings.Contains(actionStr, "buy") || strings.Contains(actionStr, "reinvest shares") {
			quantity, err := numberStrToDecimal(record[ordering["quantity"]])
			if err != nil {
				return nil, err
			}

			price, err := numberStrToDecimal(record[ordering["price"]])
			if err != nil {
				return nil, err
			}

			date, err := time.Parse("01/02/2006", record[ordering["date"]])
			if err != nil {
				return nil, err
			}

			trade := domain.Trade{
				Symbol:    record[ordering["symbol"]],
				Quantity:  quantity,
				Price:     price,
				Date:      date,
				Action:    model.TradeActionType_Buy,
				Custodian: model.CustodianType_Tda,
			}

			savepointName, err := db.AddSavepoint(tx)
			if err != nil {
				return nil, fmt.Errorf("failed to create savepoint for ProcessTdaBuyOrder: %w", err)
			}

			// still adding tda transactions
			newTrade, _, err := tiService.ProcessTdaBuyOrder(ctx, tx, trade, nil)
			if err != nil {
				if rollbackErr := db.RollbackToSavepoint(savepointName, tx); rollbackErr != nil {
					return nil, rollbackErr
				}

				if errors.As(err, &hood_errors.ErrDuplicateTrade{}) {
					fmt.Printf("skipping duplicate trade: %s\n", err.Error())
				} else {
					return nil, err
				}
			}
			orders = append(orders, *newTrade)
		} else if strings.Contains(actionStr, "sell") {
			quantity, err := numberStrToDecimal(record[ordering["quantity"]])
			if err != nil {
				return nil, err
			}

			price, err := numberStrToDecimal(record[ordering["price"]])
			if err != nil {
				return nil, err
			}

			date, err := time.Parse("01/02/2006", record[ordering["date"]])
			if err != nil {
				return nil, err
			}

			trade := domain.Trade{
				Symbol:    record[ordering["symbol"]],
				Quantity:  quantity,
				Price:     price,
				Date:      date,
				Action:    model.TradeActionType_Sell,
				Custodian: model.CustodianType_Tda,
			}

			savepointName, err := db.AddSavepoint(tx)
			if err != nil {
				return nil, fmt.Errorf("failed to create savepoint for ProcessTdaSellOrder: %w", err)
			}

			// still adding tda transactions
			newTrade, _, err := tiService.ProcessTdaSellOrder(ctx, tx, trade, nil)
			if err != nil {
				if rollbackErr := db.RollbackToSavepoint(savepointName, tx); rollbackErr != nil {
					return nil, rollbackErr
				}

				if errors.As(err, &hood_errors.ErrDuplicateTrade{}) {
					fmt.Printf("skipping duplicate trade: %s\n", err.Error())
				} else {
					return nil, err
				}
			}
			orders = append(orders, *newTrade)
		}
	}

	return orders, nil
}

func numberStrToDecimal(in string) (decimal.Decimal, error) {
	s := strings.ReplaceAll(in, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	return decimal.NewFromString(s)
}
