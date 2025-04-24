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
var counterMutex sync.Mutex

// ProcessPDF processes a PDF scan and extracts data for students' pages for the given exam.
//
// This function performs the following steps:
// 1. Clears all student pages associated with the provided exam in the database.
// 2. Loads the PDF file from the specified path.
// 3. Iterates over all pages of the PDF and processes them concurrently using goroutines.
// 4. Each page is processed by the `ProcessPage` function, which handles the extraction of relevant data.
// 5. The function waits for all pages to be processed using a WaitGroup.
//
// Parameters:
//   - scanPath: The path to the PDF file containing the scan of the answer sheets.
//   - exam: A pointer to the exam object that the pages belong to, used for database operations.
//   - db: A pointer to the GORM database object used for interacting with the database.
//   - progressChan: A channel used for sending progress updates during the processing of pages.
//   - counter: A pointer to an integer used for tracking the progress or processing count.
//
// Notes:
//   - The function uses goroutines to process each page concurrently, and a WaitGroup is used to ensure that
//     all pages are processed before returning.
//   - Errors encountered during PDF loading or database operations are logged using an error logger.
func ProcessPDF(scanPath string, exam *models.Exam, db *gorm.DB, progressChan chan string, counter *int) {
	errorLogger := logging.GetErrorLogger()

	// Vyčistenie všetkých stránok študentov pre daný test
	err := repository.ClearStudentPagesForExam(db, exam.ID)
	if err != nil {
		errorLogger.Error("Nepodarilo sa vyčistiť stránky študentov", slog.String("examID", fmt.Sprint(exam.ID)), slog.String("error", err.Error()))
		return
	}

	doc, err := fitz.New(scanPath)
	if err != nil {
		errorLogger.Error("Chyba pri načítaní PDF súboru", slog.String("file", scanPath), slog.String("error", err.Error()))
		panic(err)
	}
	totalPages := doc.NumPage()
	for pageNumber := 0; pageNumber < totalPages; pageNumber++ {
		wg.Add(1)
		go ProcessPage(doc, pageNumber, exam, db, progressChan, totalPages, counter)
	}
	wg.Wait()
}

// ProcessPage processes a single page from the provided PDF document, extracts student information,
// evaluates their answers, and updates the student record in the database.
//
// The function performs the following tasks:
// 1. Extracts the image of the specified page from the provided PDF document.
// 2. Converts the image to a grayscale matrix and applies rotation correction.
// 3. Retrieves the student associated with the page using OCR or QR code extraction.
// 4. Evaluates the student's answers from the image and stores the results in the database.
// 5. Sends progress updates to the `progressChan` channel and increments the `counter`.
// 6. Logs important steps, student information, and any errors encountered during processing.
//
// Parameters:
//   - doc: A pointer to the fitz.Document representing the loaded PDF document.
//   - pageNumber: The page number (index) to process within the document.
//   - exam: A pointer to the `models.Exam` object representing the exam details.
//   - db: A pointer to the GORM database object for database operations.
//   - progressChan: A channel used to send progress updates, such as the number of pages processed.
//   - totalPages: The total number of pages in the PDF document to track progress.
//   - counter: A pointer to an integer for counting the number of processed pages.
//
// Notes:
//   - This function uses synchronization primitives (`mutexGetId`, `mutexUpdate`, `counterMutex`) to ensure that database
//     interactions and the counter are thread-safe when processing pages concurrently.
//   - The `wg.Done()` is called to indicate the completion of processing for the current page in the goroutine.
//   - Errors are logged and the process halts further processing for the page in case of critical issues (e.g., failure to
//     extract student information or update the database).
func ProcessPage(doc *fitz.Document, pageNumber int, exam *models.Exam, db *gorm.DB, progressChan chan string, totalPages int, counter *int) {
	defer wg.Done()
	errorLogger := logging.GetErrorLogger()

	img, err := doc.Image(pageNumber)
	if err != nil {
		errorLogger.Error("Chyba pri extrahovaní obrázka z PDF stránky", slog.Int("page", pageNumber), slog.String("error", err.Error()))
		return
	}
	mat := ImageToMat(img)
	defer mat.Close()
	mat = MatToGrayscale(mat)
	mat = FixImageRotation(mat)
	mutexGetId.Lock()
	student, err := GetStudent(&mat, db, exam.ID)
	mutexGetId.Unlock()
	if err != nil {
		errorLogger.Error("Chyba pri získavaní ID študenta z databázy", "PDF strana", pageNumber, "error", err.Error())
		return
	}

	errorLogger.Info("Našiel sa študent v databáze", "studentID", student.ID, "name", student.Name)
	questionNumber, answers := EvaluateAnswers(&mat, exam.QuestionCount)
	if questionNumber == -1 {
		errorLogger.Error("Chyba pri rozpoznávaní čísiel otázok", "PDF strana", pageNumber)
		return
	}
	mutexUpdate.Lock()
	err = repository.UpdateStudentAnswers(db, student.ID, exam.ID, questionNumber, answers, pageNumber+1)
	mutexUpdate.Unlock()
	if err != nil {
		errorLogger.Error("Chyba pri aktualizácii študenta v databáze", "studentID", student.ID, "error", err.Error())
		return
	}
	errorLogger.Info("Aktualizované odpovede študenta", "studentID", student.ID, "answers", student.Answers)

	if counter != nil {
		counterMutex.Lock()
		*counter = *counter + 1
		curr := *counter
		counterMutex.Unlock()

		fmt.Println("Spracovaných", curr, "/", totalPages)
		if progressChan != nil {
			progressChan <- fmt.Sprintf("Spracovaných %d / %d", curr, totalPages)
		}
	}

}
