package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"
)

// extractSheetID extracts the Google Sheet ID from various URL formats
func extractSheetID(url string) (string, error) {
	patterns := []string{
		`/spreadsheets/d/([a-zA-Z0-9-_]+)`,
		`key=([a-zA-Z0-9-_]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	if regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString(url) {
		return url, nil
	}

	return "", fmt.Errorf("could not extract sheet ID from URL: %s", url)
}

// buildCSVExportURL creates the CSV export URL from a sheet ID
func buildCSVExportURL(sheetID string) string {
	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", sheetID)
}

// fetchCSVData fetches CSV data from a Google Sheet
func fetchCSVData(sheetURL string) ([][]string, error) {
	sheetID, err := extractSheetID(sheetURL)
	if err != nil {
		return nil, err
	}

	exportURL := buildCSVExportURL(sheetID)

	resp, err := http.Get(exportURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sheet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch sheet: HTTP %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse CSV: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// sanitizeColumnName creates a valid SQLite column name from a header
func sanitizeColumnName(header string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	name := re.ReplaceAllString(header, "_")

	re = regexp.MustCompile(`_+`)
	name = re.ReplaceAllString(name, "_")

	name = strings.Trim(name, "_")

	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "col_" + name
	}

	if name == "" {
		name = "unnamed_column"
	}

	return name
}

// createDatabase creates a SQLite database with the given schema
func createDatabase(dbPath string, headers []string) (*sql.DB, error) {
	if _, err := os.Stat(dbPath); err == nil {
		if err := os.Remove(dbPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing database: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	var columns []string
	columns = append(columns, "id INTEGER PRIMARY KEY AUTOINCREMENT")

	usedNames := make(map[string]int)
	for _, header := range headers {
		colName := sanitizeColumnName(header)

		if count, exists := usedNames[colName]; exists {
			usedNames[colName] = count + 1
			colName = fmt.Sprintf("%s_%d", colName, count+1)
		} else {
			usedNames[colName] = 1
		}

		columns = append(columns, fmt.Sprintf("\"%s\" TEXT", colName))
	}

	createSQL := fmt.Sprintf("CREATE TABLE walkthrough_data (\n  %s\n)", strings.Join(columns, ",\n  "))

	_, err = db.Exec(createSQL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}

// insertData inserts all data rows into the database
func insertData(db *sql.DB, headers []string, data [][]string) error {
	if len(data) == 0 {
		return nil
	}

	var colNames []string
	usedNames := make(map[string]int)
	for _, header := range headers {
		colName := sanitizeColumnName(header)
		if count, exists := usedNames[colName]; exists {
			usedNames[colName] = count + 1
			colName = fmt.Sprintf("%s_%d", colName, count+1)
		} else {
			usedNames[colName] = 1
		}
		colNames = append(colNames, fmt.Sprintf("\"%s\"", colName))
	}

	placeholders := make([]string, len(headers))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	insertSQL := fmt.Sprintf(
		"INSERT INTO walkthrough_data (%s) VALUES (%s)",
		strings.Join(colNames, ", "),
		strings.Join(placeholders, ", "),
	)

	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for i, row := range data {
		values := make([]interface{}, len(headers))
		for j := range headers {
			if j < len(row) {
				values[j] = row[j]
			} else {
				values[j] = ""
			}
		}

		_, err := stmt.Exec(values...)
		if err != nil {
			return fmt.Errorf("failed to insert row %d: %w", i+1, err)
		}
	}

	return nil
}

// ConvertSheetToSQLite is the main conversion function
func ConvertSheetToSQLite(sheetURL, dbPath string) error {
	records, err := fetchCSVData(sheetURL)
	if err != nil {
		return fmt.Errorf("failed to fetch CSV data: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no data found in the spreadsheet")
	}

	headers := records[0]
	data := records[1:]

	db, err := createDatabase(dbPath, headers)
	if err != nil {
		return err
	}
	defer db.Close()

	err = insertData(db, headers, data)
	if err != nil {
		return err
	}

	return nil
}

// ConvertSheetToSQLiteWithProgress is the conversion function with progress callback
func ConvertSheetToSQLiteWithProgress(sheetURL, dbPath string, onProgress func(status string)) error {
	onProgress("Fetching data from Google Sheet...")

	records, err := fetchCSVData(sheetURL)
	if err != nil {
		return fmt.Errorf("failed to fetch CSV data: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no data found in the spreadsheet")
	}

	headers := records[0]
	data := records[1:]

	onProgress(fmt.Sprintf("Found %d columns and %d rows", len(headers), len(data)))

	onProgress("Creating database...")
	db, err := createDatabase(dbPath, headers)
	if err != nil {
		return err
	}
	defer db.Close()

	onProgress("Inserting data...")
	err = insertData(db, headers, data)
	if err != nil {
		return err
	}

	onProgress(fmt.Sprintf("Successfully converted %d rows", len(data)))
	return nil
}
