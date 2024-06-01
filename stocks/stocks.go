package stocks

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
)

// Define the struct to hold the CSV data
type MyWatchlist struct {
    Name   string
    Symbol string
}

func GetStocks()  (companies []MyWatchlist){
	STOCKS_FILE_PATH, filepathErr := filepath.Abs("./stocks/stocks.csv")
	if filepathErr != nil {
		return nil
	}
	// Open the CSV file
    file, err := os.Open(STOCKS_FILE_PATH)
    if err != nil {
        log.Fatalf("failed to open file: %s", err)
    }
    defer file.Close()

    // Create a new CSV reader
    reader := csv.NewReader(file)

    // Read all the CSV records
    records, err := reader.ReadAll()
    if err != nil {
        log.Fatalf("failed to read CSV file: %s", err)
    }

    // Iterate through the CSV records and populate the struct
    for _, record := range records {
        company := MyWatchlist{
            Name:   record[0],
            Symbol: record[2],
        }

        companies = append(companies, company)
    }

	return companies
}