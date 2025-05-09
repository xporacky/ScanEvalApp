package scanprocessing

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/files/pdf"
	"sync"

	"ScanEvalApp/internal/logging"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
	"gorm.io/gorm"
)

var wg sync.WaitGroup
var mutexUpdate sync.Mutex
var mutexGetId sync.Mutex
var counterMutex sync.Mutex

type FailedPages struct {
	mu   sync.Mutex
	data map[uint][]int
}

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
//   - hadFailures: A pointer to a boolean that will be set to true if any pages failed to process.
//
// Notes:
//   - The function uses goroutines to process each page concurrently, and a WaitGroup to ensure
//     all processing is complete before proceeding.
//   - Pages that fail due to extraction, recognition, or database issues are recorded in `failedPagesMap`.
//   - At the end of processing, all failed pages are exported to a separate PDF file using `ExportFailedPagesToPDF`
//     for further inspection or manual correction.
//   - Errors encountered during PDF loading or database operations are logged using the error logger.
func ProcessPDF(scanPath string, exam *models.Exam, db *gorm.DB, progressChan chan string, counter *int, hadFailures *bool) {
	errorLogger := logging.GetErrorLogger()

	// Vyčistenie všetkých stránok študentov pre daný test
	err := repository.ClearStudentForExam(db, exam.ID)
	if err != nil {
		errorLogger.Error("Nepodarilo sa vyčistiť stránky študentov", slog.String("examID", fmt.Sprint(exam.ID)), slog.String("error", err.Error()))
		return
	}

	failedPages := &FailedPages{
		data: make(map[uint][]int),
	}

	safeTitle := common.SanitizeFilename(exam.Title)
	fileName := fmt.Sprintf("scan_%s_%d.pdf", safeTitle, exam.ID)
	if err := os.MkdirAll(common.GLOBAL_TEMP_SCAN, 0755); err != nil {
		errorLogger.Error("Nepodarilo sa vytvoriť cieľový adresár:", slog.String("error", err.Error()))
		return
	}
	destPath := filepath.Join(common.GLOBAL_TEMP_SCAN, fileName)
	err = copyFile(scanPath, destPath)
	if err != nil {
		errorLogger.Error("Chyba pri kopírovaní súboru:", slog.String("error", err.Error()))
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
		go ProcessPage(doc, pageNumber, exam, db, progressChan, totalPages, counter, failedPages)
	}
	wg.Wait()

	if len(failedPages.data) > 0 {
		*hadFailures = true
	}

	for examID, pages := range failedPages.data {
		safeTitle := common.SanitizeFilename(exam.Title)
		err := pdf.ExportFailedPagesToPDF(safeTitle, examID, pages, scanPath)
		if err != nil {
			errorLogger.Error("Nepodarilo sa exportovat PDF s chybnymi stranami", slog.String("examID", fmt.Sprint(exam.ID)), slog.String("error", err.Error()))
			return
		}
	}
}

// ProcessPage processes a single page from the provided PDF document, extracts student information,
// evaluates their answers, and updates the student record in the database.
//
// The function performs the following tasks:
//  1. Extracts the image of the specified page from the PDF document.
//  2. Converts the image to a grayscale matrix and applies rotation correction.
//  3. Retrieves the student associated with the page using OCR or QR code extraction.
//  4. Evaluates the student's answers from the image and stores the results in the database.
//  5. If any step fails (e.g., image extraction, student identification, answer recognition, or DB update),
//     the page number is recorded in `failedPagesMap`.
//  6. Sends progress updates to the `progressChan` and increments the shared `counter`.
//  7. Signals completion to the parent WaitGroup.
//
// Parameters:
//   - doc: A pointer to the fitz.Document representing the loaded PDF document.
//   - pageNumber: The page index (zero-based) to process within the document.
//   - exam: A pointer to the `models.Exam` object representing the exam details.
//   - db: A pointer to the GORM database object for database operations.
//   - progressChan: A channel used to send progress updates, such as the number of pages processed.
//   - totalPages: The total number of pages in the PDF document.
//   - counter: A pointer to an integer for counting the number of successfully processed pages.
//
// Notes:
//   - The function uses synchronization primitives (`mutexGetId`, `mutexUpdate`, `counterMutex`) to
//     ensure that concurrent access to shared resources is safe.
//   - The global `failedPagesMap` is updated in a thread-safe manner when a page fails processing.
//   - Exporting of failed pages is handled later in `ProcessPDF`, not here.
func ProcessPage(doc *fitz.Document, pageNumber int, exam *models.Exam, db *gorm.DB, progressChan chan string, totalPages int, counter *int, failedPages *FailedPages) {
	defer wg.Done()
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	img, err := doc.Image(pageNumber)
	if err != nil {
		errorLogger.Error("Chyba pri extrahovaní obrázka z PDF stránky", slog.Int("page", pageNumber), slog.String("error", err.Error()))
		AddFailedPage(failedPages, exam.ID, pageNumber)
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
		AddFailedPage(failedPages, exam.ID, pageNumber)
		return
	}

	logger.Info("Našiel sa študent v databáze", "studentID", student.ID, "name", student.Name)
	questionNumber, answers := EvaluateAnswers(&mat, exam.QuestionCount)

	if len(answers) == 0 {
		errorLogger.Error("Chyba pri rozpoznávaní odpovedí - žiadne odpovede detekované", "PDF strana", pageNumber+1)
		// Gather pageNumbers to map
		AddFailedPage(failedPages, exam.ID, pageNumber)
		return
	}

	if questionNumber == common.QUESTION_NUMBER_NOT_FOUND {
		errorLogger.Error("Chyba pri rozpoznávaní čísiel otázok - ziadna otazka detekovana", "PDF strana", pageNumber+1)
		// Gather pageNumbers to map
		AddFailedPage(failedPages, exam.ID, pageNumber)
		return
	} else if ((questionNumber + 1) % NUMBER_OF_QUESTIONS_PER_PAGE) != 0 {
		errorLogger.Error("Chyba pri rozpoznávaní čísiel otázok - menej otazok nez pocet", "PDF strana", pageNumber+1)
		// fmt.Printf("questionNumber %d %% len(answers) %d - strana: %d\n", questionNumber+1, len(answers), pageNumber+1)
		// Gather pageNumbers to map
		AddFailedPage(failedPages, exam.ID, pageNumber)
		return
	}

	mutexUpdate.Lock()
	err = repository.UpdateStudentAnswers(db, student.ID, exam.ID, questionNumber, answers, pageNumber+1)
	mutexUpdate.Unlock()

	if err != nil {
		errorLogger.Error("Chyba pri aktualizácii študenta v databáze", "studentID", student.ID, "error", err.Error())
		return
	}

	logger.Info("Aktualizované odpovede študenta", "studentID", student.ID, "answers", student.Answers)

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

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	err = destinationFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
