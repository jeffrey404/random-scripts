# Walkthrough Conversion

Converts data from a publicly viewable Google Sheet to a SQLite database.

## Features

- Fetches CSV data directly from Google Sheets (no API key required)
- Auto-generates SQLite schema from spreadsheet headers
- Sanitizes column names for SQLite compatibility
- Handles duplicate column names automatically
- Native Windows GUI with file browser dialog
- CLI version for automation and scripting

## Requirements

- Windows (GUI version)
- The Google Sheet must be publicly viewable (Share > Anyone with the link can view)

## Usage

### GUI Version

Run `walkthrough-gui.exe`:

1. Paste the Google Sheet URL
2. Click **Browse...** to select where to save the database
3. Click **Convert**

Progress and status messages are displayed during conversion.

### CLI Version

```
walkthrough-cli.exe <google-sheet-url> <output.db>
```

Example:
```
walkthrough-cli.exe "https://docs.google.com/spreadsheets/d/1NtBmGFtRKagrjgv2nd3Cg0ZRs3aEWeYUdeP8iSew_tU/edit" output.db
```

## Building from Source

Requires Go 1.21 or later.

```bash
# Build CLI version
go build -o walkthrough-cli.exe .

# Build GUI version (native Windows)
go install github.com/akavel/rsrc@latest
rsrc -manifest walkthrough-gui.manifest -o rsrc.syso
go build -tags gui -ldflags "-H windowsgui" -o walkthrough-gui.exe .
```

## Database Schema

The tool creates a single table named `walkthrough_data` with:

- `id` - Auto-incrementing primary key (INTEGER)
- One TEXT column for each spreadsheet header

Column name sanitization:
- Special characters replaced with underscores
- Consecutive underscores collapsed
- Leading/trailing underscores removed
- Columns starting with numbers prefixed with `col_`
- Duplicate names numbered (e.g., `column`, `column_2`, `column_3`)

## Project Structure

```
walkthrough-conversion/
├── converter.go              # Core conversion logic
├── main.go                   # CLI entry point
├── main_gui.go               # Native Windows GUI
├── walkthrough-gui.manifest  # Windows application manifest
├── go.mod                    # Go module definition
└── go.sum                    # Dependency checksums
```
