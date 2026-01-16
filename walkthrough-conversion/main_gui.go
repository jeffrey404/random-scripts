//go:build gui

package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var mainWindow *walk.MainWindow
	var sheetURLEdit *walk.LineEdit
	var dbPathEdit *walk.LineEdit
	var statusLabel *walk.TextLabel
	var convertBtn *walk.PushButton
	var progressBar *walk.ProgressBar

	MainWindow{
		AssignTo: &mainWindow,
		Title:    "Walkthrough Conversion",
		MinSize:  Size{Width: 500, Height: 300},
		Size:     Size{Width: 550, Height: 320},
		Layout:   VBox{MarginsZero: false},
		Children: []Widget{
			Label{Text: "Google Sheet URL:"},
			LineEdit{
				AssignTo:    &sheetURLEdit,
				ToolTipText: "Paste the full URL of a publicly viewable Google Sheet",
			},
			VSpacer{Size: 10},
			Label{Text: "Destination SQLite Database:"},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					LineEdit{
						AssignTo:    &dbPathEdit,
						ToolTipText: "Path where the SQLite database will be saved",
					},
					PushButton{
						Text:    "Browse...",
						MaxSize: Size{Width: 80},
						OnClicked: func() {
							dlg := new(walk.FileDialog)
							dlg.Title = "Save SQLite Database"
							dlg.Filter = "SQLite Database (*.db)|*.db|All Files (*.*)|*.*"

							if ok, _ := dlg.ShowSave(mainWindow); ok {
								path := dlg.FilePath
								// Ensure .db extension
								if len(path) < 3 || path[len(path)-3:] != ".db" {
									path = path + ".db"
								}
								dbPathEdit.SetText(path)
							}
						},
					},
				},
			},
			VSpacer{Size: 20},
			PushButton{
				AssignTo: &convertBtn,
				Text:     "Convert",
				OnClicked: func() {
					sheetURL := sheetURLEdit.Text()
					dbPath := dbPathEdit.Text()

					if sheetURL == "" {
						walk.MsgBox(mainWindow, "Error", "Please enter a Google Sheet URL", walk.MsgBoxIconError)
						return
					}
					if dbPath == "" {
						walk.MsgBox(mainWindow, "Error", "Please select a destination database file", walk.MsgBoxIconError)
						return
					}

					convertBtn.SetEnabled(false)
					progressBar.SetVisible(true)
					progressBar.SetMarqueeMode(true)
					statusLabel.SetText("Converting...")

					go func() {
						err := ConvertSheetToSQLiteWithProgress(sheetURL, dbPath, func(status string) {
							mainWindow.Synchronize(func() {
								statusLabel.SetText(status)
							})
						})

						mainWindow.Synchronize(func() {
							progressBar.SetMarqueeMode(false)
							progressBar.SetVisible(false)
							convertBtn.SetEnabled(true)

							if err != nil {
								statusLabel.SetText("Error: " + err.Error())
								walk.MsgBox(mainWindow, "Error", err.Error(), walk.MsgBoxIconError)
							} else {
								statusLabel.SetText("Conversion completed successfully!")
								walk.MsgBox(mainWindow, "Success", "Data has been converted to SQLite database.", walk.MsgBoxIconInformation)
							}
						})
					}()
				},
			},
			VSpacer{Size: 10},
			TextLabel{
				AssignTo: &statusLabel,
				Text:     "Ready",
			},
			ProgressBar{
				AssignTo: &progressBar,
				Visible:  false,
			},
		},
	}.Run()
}
