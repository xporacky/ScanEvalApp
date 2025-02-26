package scanprocessing

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"

	"ScanEvalApp/internal/logging"
	"log/slog"

	"github.com/gen2brain/go-fitz"
	"gorm.io/gorm"
)

// Process PDF
func ProcessPDF(scanPath string, exam *models.Exam, db *gorm.DB) {
	errorLogger := logging.GetErrorLogger()
	doc, err := fitz.New(scanPath)
	if err != nil {
		errorLogger.Error("Chyba pri načítaní PDF súboru", slog.String("file", scanPath), slog.String("error", err.Error()))
		panic(err)
	}
	for n := 0; n < doc.NumPage(); n++ {
		ProcessPage(doc, n, exam, db)
	}
}

func ProcessPage(doc *fitz.Document, n int, exam *models.Exam, db *gorm.DB) {
	errorLogger := logging.GetErrorLogger()

	img, err := doc.Image(n)
	if err != nil {
		errorLogger.Error("Chyba pri extrahovaní obrázka z PDF stránky", slog.Int("page", n), slog.String("error", err.Error()))
		panic(err)
	}
	mat := ImageToMat(img)
	mat = MatToGrayscale(mat)
	mat = FixImageRotation(mat)
	student, err := GetStudent(&mat, db, exam.ID)
	if err != nil {
		errorLogger.Error("Chyba pri získavaní ID študenta z databázy", "PDF strana", n, "error", err.Error())
		return
	}
	errorLogger.Info("Našiel sa študent v databáze", "studentID", student.ID, "name", student.Name)
	EvaluateAnswers(&mat, exam.QuestionCount, student)
	err = repository.UpdateStudent(db, student)
	if err != nil {
		errorLogger.Error("Chyba pri aktualizácii študenta v databáze", "studentID", student.ID, "error", err.Error())
		return
	}
	errorLogger.Info("Aktualizované odpovede študenta", "studentID", student.ID, "answers", student.Answers)
	//SaveMat("", mat)
	defer mat.Close()
	//println(student.Answers)
	//ShowMat(mat)
	//return
}
