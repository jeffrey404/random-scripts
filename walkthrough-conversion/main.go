//go:build !gui

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: walkthrough-conversion <google-sheet-url> <output.db>")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  walkthrough-conversion https://docs.google.com/spreadsheets/d/xxx/edit output.db")
		os.Exit(1)
	}

	sheetURL := os.Args[1]
	dbPath := os.Args[2]

	err := ConvertSheetToSQLiteWithProgress(sheetURL, dbPath, func(status string) {
		fmt.Println(status)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
