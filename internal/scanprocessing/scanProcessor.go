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
	"runtime"

	"github.com/gen2brain/go-fitz"
	"gorm.io/gorm"
)

var mutexUpdate sync.Mutex
var mutexGetId sync.Mutex
var counterMutex sync.Mutex

type FailedPageInfo struct {
	PageNumber            int    `json:"pageNumber"`
	Reason                string `json:"reason"`
	ExamTitle             string `json:"examTitle"`
	ExamDate              string `json:"examDate"`
	ExamTime              string `json:"examTime"`
	Room                  string `json:"room"`
	ExtractedAnswers      string `json:"extractedAnswers"`
	UnrecognizedQuestions []int  `json:"unrecognizedQuestions"` // Question numbers that had 'x' (unrecognized)
	DetailedReason        string `json:"detailedReason"`
}

type FailedPages struct {
	mu   sync.Mutex
	data map[uint][]FailedPageInfo
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
func ProcessPDF(scanPath string, exam *models.Exam, db *gorm.DB, progressChan chan string, counter *int, hadFailures *bool) ([]FailedPageInfo, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Nevymažeme študentov - umožňuje inkrementálne vyhodnotenie po častiach
	// UpdateStudentAnswers teraz inteligentne zvládne duplicitné stránky

	failedPages := &FailedPages{
		data: make(map[uint][]FailedPageInfo),
	}

	safeTitle := common.SanitizeFilename(exam.Title)
	fileName := fmt.Sprintf("scan_%s_%d.pdf", safeTitle, exam.ID)
	if err := os.MkdirAll(common.GLOBAL_TEMP_SCAN, 0755); err != nil {
		errorLogger.Error("Nepodarilo sa vytvoriť cieľový adresár:", slog.String("error", err.Error()))
		return nil, err
	}
	destPath := filepath.Join(common.GLOBAL_TEMP_SCAN, fileName)
	err := copyFile(scanPath, destPath)
	if err != nil {
		errorLogger.Error("Chyba pri kopírovaní súboru:", slog.String("error", err.Error()))
		return nil, err
	}

	doc, err := fitz.New(scanPath)
	if err != nil {
		errorLogger.Error("Chyba pri načítaní PDF súboru", slog.String("file", scanPath), slog.String("error", err.Error()))
		return nil, err
	}
	defer doc.Close()
	var wg sync.WaitGroup
	var docMutex sync.Mutex
	numWorkers := runtime.NumCPU()
	if numWorkers > 6 {
		numWorkers = 6
	}
	sem := make(chan struct{}, numWorkers)
	totalPages := doc.NumPage()
	for pageNumber := 0; pageNumber < totalPages; pageNumber++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(pn int) {
			defer func() { <-sem }()
			ProcessPage(&wg, &docMutex, doc, pn, exam, db, progressChan, totalPages, counter, failedPages)
		}(pageNumber)
	}
	wg.Wait()

	if len(failedPages.data) > 0 {
		*hadFailures = true
		logger.Info("=== ZHRNUTIE ZLYHANÝCH STRÁN ===")
	}

	// Collect all failed pages for return
	var allFailedPages []FailedPageInfo

	for examID, pageInfos := range failedPages.data {
		safeTitle := common.SanitizeFilename(exam.Title)

		// Add to return list
		allFailedPages = append(allFailedPages, pageInfos...)

		// Rozdelenie podľa dôvodov
		reasonMap := make(map[string][]int)
		pageNumbers := make([]int, len(pageInfos))

		for i, info := range pageInfos {
			pageNumbers[i] = info.PageNumber
			reasonMap[info.Reason] = append(reasonMap[info.Reason], info.PageNumber+1) // +1 pre ľudsky čitateľné číslo
		}

		// Logovanie rozdelené podľa dôvodov
		logger.Info("Zlyhané strany pre exam", slog.Int("examID", int(examID)), slog.Int("total", len(pageInfos)))
		for reason, pages := range reasonMap {
			logger.Info("Dôvod zlyhania",
				slog.String("reason", reason),
				slog.Any("pages", pages),
				slog.Int("count", len(pages)))
		}

		err := pdf.ExportFailedPagesToPDF(safeTitle, examID, pageNumbers, scanPath)
		if err != nil {
			errorLogger.Error("Nepodarilo sa exportovat PDF s chybnymi stranami", slog.String("examID", fmt.Sprint(exam.ID)), slog.String("error", err.Error()))
			return allFailedPages, err
		}
		logger.Info("=================================")
	}

	return allFailedPages, nil
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
func ProcessPage(wg *sync.WaitGroup, docMutex *sync.Mutex, doc *fitz.Document, pageNumber int, exam *models.Exam, db *gorm.DB, progressChan chan string, totalPages int, counter *int, failedPages *FailedPages) {
	defer wg.Done()
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()
	defer reportPageDone(progressChan, counter, totalPages)
	defer func() {
		if r := recover(); r != nil {
			errorLogger.Error("Panic v ProcessPage - strana preskočená", "strana", pageNumber+1, "recover", fmt.Sprint(r))
			AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
				PageNumber:     pageNumber,
				Reason:         "PANIC",
				ExamTitle:      exam.Title,
				ExamDate:       exam.Date.Format("02.01.2006"),
				ExamTime:       "",
				Room:           "",
				DetailedReason: fmt.Sprintf("Kritická chyba pri spracovaní: %v", r),
			})
		}
	}()

	docMutex.Lock()
	img, err := doc.Image(pageNumber)
	docMutex.Unlock()
	if err != nil {
		errorLogger.Error("Chyba pri extrahovaní obrázka z PDF stránky", slog.Int("page", pageNumber), slog.String("error", err.Error()))
		AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
			PageNumber:     pageNumber,
			Reason:         "IMAGE_EXTRACTION_ERROR",
			ExamTitle:      exam.Title,
			ExamDate:       exam.Date.Format("02.01.2006"),
			ExamTime:       "",
			Room:           "",
			DetailedReason: fmt.Sprintf("Chyba pri extrakcii obrázka: %s", err.Error()),
		})
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
		errorLogger.Error("Chyba pri získavaní ID študenta z databázy", "PDF strana", pageNumber+1, "error", err.Error())
		AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
			PageNumber:     pageNumber,
			Reason:         "ID_NOT_FOUND",
			ExamTitle:      exam.Title,
			ExamDate:       exam.Date.Format("02.01.2006"),
			ExamTime:       "",
			Room:           "",
			DetailedReason: fmt.Sprintf("Študent nebol nájdený: %s", err.Error()),
		})
		return
	}

	logger.Info("Našiel sa študent v databáze", "studentID", student.ID, "name", student.Name)

	// Detekcia skupiny/podskupiny z hlavičky (len pre multiday testy)
	//	if exam.IsMultiDay && student.Subgroup == "" {
	if exam.IsMultiDay {
		if groupCode := DetectGroupFromHeader(&mat, pageNumber+1); groupCode != "" {
			if err := repository.UpdateStudentSubgroup(db, student.ID, exam.ID, groupCode); err != nil {
				errorLogger.Error("Chyba pri ukladaní skupiny študenta", "studentID", student.ID, "groupCode", groupCode, "error", err.Error())
			} else {
				logger.Info("Priradená skupina študentovi", "studentID", student.ID, "groupCode", groupCode)
				student.Subgroup = groupCode
			}
		} else {
			errorLogger.Error("Nepodarilo sa rozpoznať skupinu z hlavičky", "PDF strana", pageNumber+1, "studentID", student.ID)
			AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
				PageNumber:     pageNumber,
				Reason:         "GROUP_NOT_RECOGNIZED",
				ExamTitle:      exam.Title,
				ExamDate:       student.ExamDate.Format("02.01.2006"),
				ExamTime:       student.ExamTime,
				Room:           student.Room,
				DetailedReason: "Nepodarilo sa rozpoznať skupinu/podskupinu z hlavičky (multiday test)",
			})
			return
		}
	}

	//questionNumber, answers := EvaluateAnswers(&mat, exam.QuestionCount)
	choices := exam.OptionCount
	if choices == 0 {
		choices = NUMBER_OF_CHOICES
	}
	questionNumber, answers := EvaluateAnswers(&mat, exam.QuestionCount, choices)

	if len(answers) == 0 {
		errorLogger.Error("Chyba pri rozpoznávaní odpovedí - žiadne odpovede detekované", "PDF strana", pageNumber+1)
		// Gather pageNumbers to map
		AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
			PageNumber:       pageNumber,
			Reason:           "NO_ANSWERS_DETECTED",
			ExamTitle:        exam.Title,
			ExamDate:         student.ExamDate.Format("02.01.2006"),
			ExamTime:         student.ExamTime,
			Room:             student.Room,
			ExtractedAnswers: "",
			DetailedReason:   "Na stránke neboli detekované žiadne odpovede",
		})
		return
	}

	if questionNumber == common.QUESTION_NUMBER_NOT_FOUND {
		errorLogger.Error("Chyba pri rozpoznávaní čísiel otázok - ziadna otazka detekovana", "PDF strana", pageNumber+1)
		// Gather pageNumbers to map
		AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
			PageNumber:       pageNumber,
			Reason:           "NO_QUESTION_NUMBERS",
			ExamTitle:        exam.Title,
			ExamDate:         student.ExamDate.Format("02.01.2006"),
			ExamTime:         student.ExamTime,
			Room:             student.Room,
			ExtractedAnswers: string(answers),
			DetailedReason:   "Neboli rozpoznané čísla otázok na stránke",
		})
		return
	}

	// Check for unrecognized answers during scanning (GetAnswer returned 'x')
	var unrecognizedQuestions []int
	startQuestionNum := (questionNumber - len(answers)) + 1
	for i, ans := range answers {
		// Check for 'x' or null character (rune(0))
		// Note: '0' comes from CSV initialization, not from GetAnswer
		if ans == 'x' || ans == rune(0) {
			unrecognizedQuestions = append(unrecognizedQuestions, startQuestionNum+i+1) // +1 for 1-based indexing
		}
	}

	// If there are unrecognized answers during scanning, log and replace with 'X'
	if len(unrecognizedQuestions) > 0 {
		// Replace unrecognized characters with uppercase 'X' for database storage
		for i := range answers {
			if answers[i] == 'x' || answers[i] == rune(0) {
				answers[i] = 'X'
			}
		}

		errorLogger.Warn("Stránka obsahuje nerozpoznané odpovede počas skenovania",
			"PDF strana", pageNumber+1,
			"studentID", student.ID,
			"študent", student.Name,
			"nerozpoznané otázky", unrecognizedQuestions,
			"extrahované odpovede", string(answers))
	}

	// Validácia počtu otázok: strana musí mať buď plných 20 otázok, alebo byť poslednou stranou s menej otázkami
	questionsOnPage := (questionNumber + 1) % NUMBER_OF_QUESTIONS_PER_PAGE
	if questionsOnPage == 0 {
		questionsOnPage = NUMBER_OF_QUESTIONS_PER_PAGE
	}

	// Skontrolujeme či je to buď plná strana (20 otázok) alebo posledná strana testu
	isFullPage := questionsOnPage == NUMBER_OF_QUESTIONS_PER_PAGE
	isLastPage := (questionNumber + 1) == exam.QuestionCount

	if !isFullPage && !isLastPage {
		errorLogger.Error("Chyba pri rozpoznávaní čísiel otázok - neočakávaný počet otázok na strane",
			"PDF strana", pageNumber+1,
			"posledná otázka", questionNumber+1,
			"otázok na strane", questionsOnPage,
			"očakávaných", exam.QuestionCount)
		AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
			PageNumber:       pageNumber,
			Reason:           "INVALID_QUESTION_COUNT",
			ExamTitle:        exam.Title,
			ExamDate:         student.ExamDate.Format("02.01.2006"),
			ExamTime:         student.ExamTime,
			Room:             student.Room,
			ExtractedAnswers: string(answers),
			DetailedReason:   fmt.Sprintf("Neočakávaný počet otázok: %d na strane (očakávaných %d)", questionsOnPage, exam.QuestionCount),
		})
		return
	}

	mutexUpdate.Lock()
	//err = repository.UpdateStudentAnswers(db, student.ID, exam.ID, questionNumber, answers, pageNumber+1)
	err = repository.UpdateStudentAnswers(db, student.ID, exam.ID, questionNumber, answers, pageNumber+1)
	mutexUpdate.Unlock()

	if err != nil {
		errorLogger.Error("Chyba pri aktualizácii študenta v databáze", "studentID", student.ID, "error", err.Error())
		AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
			PageNumber:       pageNumber,
			Reason:           "DB_UPDATE_ERROR",
			ExamTitle:        exam.Title,
			ExamDate:         student.ExamDate.Format("02.01.2006"),
			ExamTime:         student.ExamTime,
			Room:             student.Room,
			ExtractedAnswers: string(answers),
			DetailedReason:   fmt.Sprintf("Chyba pri ukladaní do databázy: %s", err.Error()),
		})
		return
	}

	// Reload student from DB to get final answers after update
	updatedStudent, err := repository.GetStudentById(db, student.ID, exam.ID)
	if err != nil {
		errorLogger.Error("Chyba pri načítaní študenta po update", "studentID", student.ID, "error", err.Error())
		return
	}

	logger.Info("Aktualizované odpovede študenta", "studentID", updatedStudent.ID, "answers", updatedStudent.Answers)

	// Check if final student answers still contain '0' or 'X' (incomplete/unrecognized)
	var incompleteQuestions []int
	finalAnswers := []rune(updatedStudent.Answers)
	for i, ans := range finalAnswers {
		if ans == '0' || ans == 'X' {
			incompleteQuestions = append(incompleteQuestions, i+1) // +1 for 1-based indexing
		}
	}

	// If there are incomplete answers after update, add to failed pages
	if len(incompleteQuestions) > 0 {
		errorLogger.Warn("Študent má neúplné odpovede po spracovaní strany",
			"PDF strana", pageNumber+1,
			"studentID", updatedStudent.ID,
			"študent", updatedStudent.Name,
			"neúplné otázky", incompleteQuestions,
			"finálne odpovede", updatedStudent.Answers)

		AddFailedPageDetailed(failedPages, exam.ID, FailedPageInfo{
			PageNumber:            pageNumber,
			Reason:                "PARTIAL_RECOGNITION",
			ExamTitle:             exam.Title,
			ExamDate:              updatedStudent.ExamDate.Format("02.01.2006"),
			ExamTime:              updatedStudent.ExamTime,
			Room:                  updatedStudent.Room,
			ExtractedAnswers:      updatedStudent.Answers,
			UnrecognizedQuestions: incompleteQuestions,
			DetailedReason:        fmt.Sprintf("Po spracovaní strany zostali neúplné odpovede v otázkách %v (označené ako 0 alebo X)", incompleteQuestions),
		})
	}

}

func reportPageDone(progressChan chan string, counter *int, totalPages int) {
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
