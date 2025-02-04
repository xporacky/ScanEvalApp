package scanprocessing

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"os"
	"path/filepath"

	"ScanEvalApp/internal/logging"
	"log/slog"

	"github.com/gen2brain/go-fitz"
	"gorm.io/gorm"
)

// Process PDF
func ProcessPDF(scanPath string, outputPath string, test *models.Test, db *gorm.DB) {
	errorLogger := logging.GetErrorLogger()

	var files []string

	err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".pdf" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		errorLogger.Error("Chyba pri prechádzaní adresára", slog.String("error", err.Error()))
		panic(err)
	}
	for _, file := range files {
		doc, err := fitz.New(file)
		if err != nil {
			errorLogger.Error("Chyba pri načítaní PDF súboru", slog.String("file", file), slog.String("error", err.Error()))
			panic(err)
		}
		//folder := strings.TrimSuffix(path.Base(file), filepath.Ext(path.Base(file)))

		// Extract pages as images
		for n := 0; n < doc.NumPage(); n++ {
			ProcessPage(doc, n, test, db)
		}
	}
}

func ProcessPage(doc *fitz.Document, n int, test *models.Test, db *gorm.DB) {
	errorLogger := logging.GetErrorLogger()

	img, err := doc.Image(n)
	if err != nil {
		errorLogger.Error("Chyba pri extrahovaní obrázka z PDF stránky", slog.Int("page", n), slog.String("error", err.Error()))
		panic(err)
	}
	mat := ImageToMat(img)
	mat = MatToGrayscale(mat)
	mat = FixImageRotation(mat)
	studentID, err := GetStudentID(&mat)
	if err != nil {
		errorLogger.Error("Chyba pri získavaní ID študenta", slog.String("error", err.Error()))
		return
	}
	student, err := repository.GetStudent(db, uint(studentID), test.ID)
	if err != nil {
		errorLogger.Error("Chyba pri získavaní ID študenta z databázy", "studentID", studentID, "error", err.Error())
		return
	}
	errorLogger.Info("Našiel sa študent v databáze", "studentID", student.ID, "name", student.Name)
	//path := filepath.Join("./"+outputPath+"/", fmt.Sprintf("%s-image-%05d.png", folder, n)) //na testovanie zatial takto
	EvaluateAnswers(&mat, test.QuestionCount, student)
	err = repository.UpdateStudent(db, student)
	if err != nil {
		errorLogger.Error("Chyba pri aktualizácii študenta v databáze", "studentID", student.ID, "error", err.Error())
		return
	}
	errorLogger.Info("Aktualizované odpovede študenta", "studentID", studentID, "answers", student.Answers)
	//SaveMat("", mat)
	defer mat.Close()
	//println(student.Answers)
	//ShowMat(mat)
	//return
}
