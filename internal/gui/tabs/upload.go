package tabs

import (
	"fmt"
	"gioui.org/app"
	//"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	"io"
	"log"

	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"strings"
	"time"
	"gorm.io/gorm"
	"encoding/csv"
)

type UploadTab struct {
	button       	widget.Clickable
	explorer     	*explorer.Explorer
	selectedFile 	string

}

func NewUploadTab(w *app.Window) *UploadTab {
	return &UploadTab{
		explorer: explorer.NewExplorer(w),
	}
}

func (t *UploadTab) Layout(gtx layout.Context, th *material.Theme, db *gorm.DB) layout.Dimensions {
	// Spracovanie kliknutí na tlačidlo
	if t.button.Clicked(gtx) {
		go t.openFileDialog()
	}

	// Vykreslenie layoutu
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceEvenly,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(th, &t.button, "Vybrať súbor").Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := "Žiadny súbor nebol vybraný"
				if t.selectedFile != "" {
					text = fmt.Sprintf("Vybraný súbor: %s", t.selectedFile)
					ParseCSV(strings.NewReader(t.selectedFile), db)
					fmt.Println("volam parse")
				}
				return material.Label(th, unit.Sp(16), text).Layout(gtx)
			}),
		)
	})
}

// Funkcia na otvorenie dialógového okna na výber súboru
func (t *UploadTab) openFileDialog() {
	file, err := t.explorer.ChooseFile()
	if err != nil {
		log.Println("Chyba pri výbere súboru:", err)
		return
	}
	if file != nil {
		defer file.Close() // Nezabudni zatvoriť súbor
		b, err := io.ReadAll(file)
		if err != nil {
			log.Println("Chyba pri čítaní súboru:", err)
			return
		}
		t.selectedFile = string(b)


	}
}

// Spracovanie eventov (pre Explorer)
func (t *UploadTab) HandleEvent(evt interface{}) { // Zmena na interface{}
	switch e := evt.(type) {
	case app.FrameEvent:
		t.explorer.ListenEvents(e)
	}
}


func ParseCSV(file io.Reader, db *gorm.DB) {
	
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Println("error1: %s", err)
		return 
	}


	fmt.Println("studenti v csv: %s", rows)

	for i, row := range rows {
		fmt.Println("som dnu for")
		if i == 0 {
			fmt.Println("hlavicka")
			continue // Preskočiť hlavičku CSV
		}
		birthDate, err := time.Parse("2006-01-02", row[2]) 
		if err != nil {
			fmt.Println("error2: %s", err)
			return
		}

		student := models.Student{
			Name:               row[0],
			Surname:            row[1],
			BirthDate:          birthDate,
			RegistrationNumber: row[3],
			Room:               row[4],
			TestID:             2,					//TODO:preroobilt!!!!!!!!!!!!!!!!!!!!! 
		}
		if err := repository.CreateStudent(db, &student); err != nil {
			fmt.Println("error3: ", err)
			return 
		}else{
			fmt.Println("student pridany: %s", student)
		}
	}
}



