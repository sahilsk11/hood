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

	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func determineColumnOrder(headerRow []string) (map[string]int, error) {
	requiredColumns := []string{
		"date",
		"description",
		"transaction_id",
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

func ParseTdaTransactionFile(ctx context.Context, tx *sql.Tx, csvFileName string, tiService TradeIngestionService) ([]model.Trade, error) {
	f, err := os.Open(csvFileName)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	revisedFile := strings.ReplaceAll(string(bytes), "\n***END OF FILE***", "")
	err = os.WriteFile(csvFileName, []byte(revisedFile), 0644)
	if err != nil {
		return nil, err
	}

	f, err = os.Open(csvFileName)
	if err != nil {
		return nil, err
	}

	csvFile := csv.NewReader(f)
	records, err := csvFile.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv file: %w", err)
	}

	orders := []model.Trade{}

	ordering, err := determineColumnOrder(records[0])
	if err != nil {
		return nil, err
	}

	for _, record := range records[1:] {
		descriptionStr := strings.ToLower(record[ordering["description"]])
		if strings.Contains(descriptionStr, "bought") {
			quantity, err := decimal.NewFromString(record[ordering["quantity"]])
			if err != nil {
				return nil, err
			}

			price, err := decimal.NewFromString(record[ordering["price"]])
			if err != nil {
				return nil, err
			}

			date, err := time.Parse("01/02/2006", record[ordering["date"]])
			if err != nil {
				return nil, err
			}

			transactionID, err := strconv.ParseInt(record[ordering["transaction_id"]], 10, 64)
			if err != nil {
				return nil, err
			}

			input := ProcessTdaBuyOrderInput{
				Symbol:           record[ordering["symbol"]],
				TdaTransactionID: transactionID,
				Quantity:         quantity,
				CostBasis:        price,
				Date:             date,
			}

			savepointName, err := db.AddSavepoint(tx)
			if err != nil {
				return nil, fmt.Errorf("failed to create savepoint for ProcessTdaBuyOrder: %w", err)
			}

			_, _, err = tiService.ProcessTdaBuyOrder(ctx, tx, input)
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
		}
	}

	return orders, nil
}
