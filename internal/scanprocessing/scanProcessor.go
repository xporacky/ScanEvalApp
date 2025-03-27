package scanprocessing

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"sync"

	"ScanEvalApp/internal/logging"
	"fmt"
	"log/slog"

	"github.com/gen2brain/go-fitz"
	"gorm.io/gorm"
)

var wg sync.WaitGroup
var mutexUpdate sync.Mutex
var mutexGetId sync.Mutex

// Process PDF
func ProcessPDF(scanPath string, exam *models.Exam, db *gorm.DB, progressChan chan string) {
	errorLogger := logging.GetErrorLogger()
	doc, err := fitz.New(scanPath)
	if err != nil {
		errorLogger.Error("Chyba pri načítaní PDF súboru", slog.String("file", scanPath), slog.String("error", err.Error()))
		panic(err)
	}
	totalPages := doc.NumPage()
	for n := 0; n < totalPages; n++ {
		wg.Add(1)
		go ProcessPage(doc, n, exam, db, progressChan, totalPages)
	}
	wg.Wait()
}

func ProcessPage(doc *fitz.Document, n int, exam *models.Exam, db *gorm.DB, progressChan chan string, totalPages int) {
	defer wg.Done()
	errorLogger := logging.GetErrorLogger()

	img, err := doc.Image(n)
	if err != nil {
		errorLogger.Error("Chyba pri extrahovaní obrázka z PDF stránky", slog.Int("page", n), slog.String("error", err.Error()))
		panic(err)
	}
	mat := ImageToMat(img)
	defer mat.Close()
	mat = MatToGrayscale(mat)
	mat = FixImageRotation(mat)
	mutexGetId.Lock()
	student, err := GetStudent(&mat, db, exam.ID)
	mutexGetId.Unlock()
	if err != nil {
		errorLogger.Error("Chyba pri získavaní ID študenta z databázy", "PDF strana", n, "error", err.Error())
		return
	}
	errorLogger.Info("Našiel sa študent v databáze", "studentID", student.ID, "name", student.Name)
	questionNumber, answers := EvaluateAnswers(&mat, exam.QuestionCount)
	mutexUpdate.Lock()
	err = repository.UpdateStudentAnswers(db, student.ID, exam.ID, questionNumber, answers)
	mutexUpdate.Unlock()
	if err != nil {
		errorLogger.Error("Chyba pri aktualizácii študenta v databáze", "studentID", student.ID, "error", err.Error())
		return
	}
	errorLogger.Info("Aktualizované odpovede študenta", "studentID", student.ID, "answers", student.Answers)
	fmt.Println("Spracovaných ", n+1, "/", totalPages)
	if progressChan != nil {
		progressChan <- fmt.Sprintf("Spracovaných %d / %d", n+1, totalPages)
	}
}
