# Walkthrough Conversion

Converts data from a publicly viewable Google Sheet to a SQLite database.

## Features

- Fetches CSV data directly from Google Sheets (no API key required)
- Auto-generates SQLite schema from spreadsheet headers
- Sanitizes column names for SQLite compatibility
- Handles duplicate column names automatically
- Native Windows GUI with file browser dialog
- CLI version for automation and scripting
- Progress feedback during conversion

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      User Interface                              │
├────────────────────────┬────────────────────────────────────────┤
│     CLI (main.go)      │         GUI (main_gui.go)              │
│  - Command line args   │  - Native Windows UI (walk)            │
│  - Console output      │  - File browser dialog                 │
│                        │  - Progress bar                        │
└────────────┬───────────┴──────────────────┬─────────────────────┘
             │                              │
             └──────────────┬───────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Converter (converter.go)                       │
├─────────────────────────────────────────────────────────────────┤
│  ConvertSheetToSQLite() / ConvertSheetToSQLiteWithProgress()    │
│                            │                                     │
│         ┌──────────────────┼──────────────────┐                 │
│         ▼                  ▼                  ▼                 │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐           │
│  │ fetchCSVData│   │createDatabase│   │ insertData  │           │
│  │             │   │             │   │             │           │
│  │ - Extract ID│   │ - Sanitize  │   │ - Prepare   │           │
│  │ - HTTP GET  │   │   columns   │   │   statement │           │
│  │ - Parse CSV │   │ - Create    │   │ - Batch     │           │
│  │             │   │   table     │   │   insert    │           │
│  └─────────────┘   └─────────────┘   └─────────────┘           │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    External Services                             │
├─────────────────────────────────────────────────────────────────┤
│  Google Sheets CSV Export API    │    SQLite Database (file)    │
│  GET /spreadsheets/d/{id}/       │    - modernc.org/sqlite      │
│      export?format=csv           │    - Pure Go implementation  │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ Google Sheet │────▶│  CSV Export  │────▶│  Parse CSV   │────▶│   SQLite DB  │
│     URL      │     │   HTTP GET   │     │   Records    │     │    File      │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
      │                    │                    │                    │
      │                    │                    │                    │
      ▼                    ▼                    ▼                    ▼
 Extract Sheet ID    Fetch CSV data      Headers → Schema      Insert rows
 from URL patterns   from Google API     Data → Records        with prepared
                                                                statements
```

## Requirements

- **Windows** (GUI version uses native Windows controls)
- **Google Sheet** must be publicly viewable (Share > Anyone with the link can view)

## Installation

### Pre-built Binaries

Download from the releases page or build from source.

### Build from Source

Requires Go 1.21 or later.

```bash
# Clone the repository
git clone https://github.com/jeffrey404/random-scripts.git
cd random-scripts/walkthrough-conversion

# Build CLI version
go build -o walkthrough-cli.exe .

# Build GUI version (native Windows)
go install github.com/akavel/rsrc@latest
rsrc -manifest walkthrough-gui.manifest -o rsrc.syso
go build -tags gui -ldflags "-H windowsgui" -o walkthrough-gui.exe .
```

## Usage

### GUI Version

Run `walkthrough-gui.exe`:

1. Paste the Google Sheet URL in the text field
2. Click **Browse...** to open a file picker and select where to save the database
3. Click **Convert**

Progress and status messages are displayed during conversion.

### CLI Version

```
walkthrough-cli.exe <google-sheet-url> <output.db>
```

**Arguments:**

| Argument | Description | Required |
|----------|-------------|----------|
| `google-sheet-url` | Full URL to a publicly viewable Google Sheet | Yes |
| `output.db` | Path where the SQLite database will be created | Yes |

**Example:**

```bash
walkthrough-cli.exe "https://docs.google.com/spreadsheets/d/1NtBmGFtRKagrjgv2nd3Cg0ZRs3aEWeYUdeP8iSew_tU/edit" output.db
```

**Output:**

```
Fetching data from Google Sheet...
Found 130 columns and 2184 rows
Creating database...
Inserting data...
Successfully converted 2184 rows
```

**Exit Codes:**

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Error (invalid arguments, network error, or database error) |

### Supported URL Formats

The tool accepts various Google Sheets URL formats:

```
# Standard edit URL
https://docs.google.com/spreadsheets/d/SHEET_ID/edit

# With sharing parameter
https://docs.google.com/spreadsheets/d/SHEET_ID/edit?usp=sharing

# Legacy key parameter
https://docs.google.com/spreadsheets?key=SHEET_ID

# Direct sheet ID
SHEET_ID
```

## Database Schema

### Table: `walkthrough_data`

The tool creates a single table with auto-generated columns:

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | Auto-incrementing primary key |
| `{header_1}` | TEXT | First spreadsheet column |
| `{header_2}` | TEXT | Second spreadsheet column |
| ... | TEXT | Additional columns from headers |

### Column Name Sanitization

Spreadsheet headers are converted to valid SQLite column names:

| Rule | Example |
|------|---------|
| Special characters → underscore | `User Name` → `User_Name` |
| Consecutive underscores collapsed | `A__B` → `A_B` |
| Leading/trailing underscores removed | `_Name_` → `Name` |
| Numbers at start prefixed | `123Col` → `col_123Col` |
| Empty names replaced | `` → `unnamed_column` |
| Duplicates numbered | `Name`, `Name` → `Name`, `Name_2` |

### System Tables

SQLite automatically creates:

| Table | Purpose |
|-------|---------|
| `sqlite_sequence` | Tracks AUTOINCREMENT values for guaranteed unique IDs |

## API Reference

### Core Functions

#### `ConvertSheetToSQLite(sheetURL, dbPath string) error`

Converts a Google Sheet to SQLite database.

**Parameters:**
- `sheetURL` - URL or ID of the Google Sheet
- `dbPath` - Output path for the SQLite database

**Returns:**
- `error` - nil on success, error message on failure

#### `ConvertSheetToSQLiteWithProgress(sheetURL, dbPath string, onProgress func(string)) error`

Converts with progress callback for UI updates.

**Parameters:**
- `sheetURL` - URL or ID of the Google Sheet
- `dbPath` - Output path for the SQLite database
- `onProgress` - Callback function receiving status messages

**Progress Messages:**
1. "Fetching data from Google Sheet..."
2. "Found X columns and Y rows"
3. "Creating database..."
4. "Inserting data..."
5. "Successfully converted X rows"

### Internal Functions

| Function | Description |
|----------|-------------|
| `extractSheetID(url)` | Extracts sheet ID from various URL formats |
| `buildCSVExportURL(id)` | Constructs the CSV export API URL |
| `fetchCSVData(url)` | Downloads and parses CSV data |
| `sanitizeColumnName(header)` | Converts header to valid column name |
| `createDatabase(path, headers)` | Creates SQLite database with schema |
| `insertData(db, headers, data)` | Inserts all data rows |

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `modernc.org/sqlite` | v1.44+ | Pure Go SQLite driver (no CGO) |
| `github.com/lxn/walk` | v0.0.0 | Native Windows GUI (GUI build only) |
| `github.com/lxn/win` | v0.0.0 | Windows API bindings (GUI build only) |

## Project Structure

```
walkthrough-conversion/
├── converter.go              # Core conversion logic
├── main.go                   # CLI entry point (build tag: !gui)
├── main_gui.go               # Native Windows GUI (build tag: gui)
├── walkthrough-gui.manifest  # Windows application manifest
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── .gitignore                # Git ignore rules
└── README.md                 # This documentation
```

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| "could not extract sheet ID" | Invalid URL format | Use a valid Google Sheets URL |
| "failed to fetch sheet: HTTP 404" | Sheet not found or not public | Make sheet publicly viewable |
| "failed to fetch sheet: HTTP 403" | Access denied | Check sharing permissions |
| "failed to parse CSV" | Malformed spreadsheet data | Check for special characters in data |
| "failed to create database" | Invalid path or permissions | Check write permissions |
| "no data found in the spreadsheet" | Empty spreadsheet | Add data to the sheet |

## Automation

The CLI version is suitable for automation via scripts or scheduled tasks:

**PowerShell:**
```powershell
# Daily export script
$sheetUrl = "https://docs.google.com/spreadsheets/d/YOUR_SHEET_ID/edit"
$dbPath = "C:\Data\walkthrough_$(Get-Date -Format 'yyyyMMdd').db"

& .\walkthrough-cli.exe $sheetUrl $dbPath

if ($LASTEXITCODE -eq 0) {
    Write-Host "Export successful: $dbPath"
} else {
    Write-Host "Export failed" -ForegroundColor Red
}
```

**Batch:**
```batch
@echo off
set SHEET_URL=https://docs.google.com/spreadsheets/d/YOUR_SHEET_ID/edit
set DB_PATH=C:\Data\walkthrough_%date:~-4,4%%date:~-10,2%%date:~-7,2%.db

walkthrough-cli.exe "%SHEET_URL%" "%DB_PATH%"

if %ERRORLEVEL% EQU 0 (
    echo Export successful
) else (
    echo Export failed
)
```

## License

MIT License - See repository root for details.
