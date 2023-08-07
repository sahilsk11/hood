package main

import (
	"encoding/csv"
	db "hood/internal/db/query"
	"hood/internal/prices"
	"log"
	"os"
)

func main() {
	tx, err := db.NewTx()
	if err != nil {
		log.Fatal(err)
	}
	csvFileName := "result.csv"
	records, err := loadCsv(csvFileName)
	if err != nil {
		log.Fatalf("failed to load csv: %v", err)
	}
	err = prices.UpdateFromCsv(tx, records)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func loadCsv(csvFileName string) ([][]string, error) {
	f, err := os.Open(csvFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvFile := csv.NewReader(f)
	records, err := csvFile.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}
