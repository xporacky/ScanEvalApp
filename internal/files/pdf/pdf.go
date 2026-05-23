package pdf

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/config"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/logging"
	"fmt"
	"image"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gen2brain/go-fitz"
	"gocv.io/x/gocv"
	"gorm.io/gorm"
)

// SlicePdfForStudent slices a PDF file based on the pages specified in the students record in DB.
// It uses the pdftk tool to extract specific pages from the input PDF and saves the result to an output PDF.
func SlicePdfForStudent(db *gorm.DB, studentID uint) (string, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	var student models.Student
	if err := db.First(&student, studentID).Error; err != nil {
		errorLogger.Error("Error finding student", "student_id", studentID, slog.String("error", err.Error()))
		return "", err
	}
	registrationNumber := student.RegistrationNumber

	pagesStr := student.Pages

	if pagesStr == "" {
		err := fmt.Errorf("študent (číslo registrácie: %d) nemá žiadne stránky v databáze", registrationNumber)
		logger.Info("Študent nemá žiadne stránky v DB", "registration_number", registrationNumber)
		return "", err
	}

	pageParts := strings.Split(pagesStr, "-")
	var pages []int
	for _, part := range pageParts {
		if part == "" {
			continue // skipping empty parts
		}
		pageNum, err := strconv.Atoi(part)
		if err != nil {
			errorLogger.Error("Invalid page number in Pages", slog.String("value", part), slog.String("error", err.Error()))
			return "", err
		}
		pages = append(pages, pageNum)
	}

	logger.Info("Parsed pages", "registration_number", registrationNumber, "pages", pages)
	exam, err := repository.GetExam(db, student.ExamID)
	if err != nil {
		errorLogger.Error("Error retrieving exam", "exam_id", student.ExamID, slog.String("error", err.Error()))
		return "", err
	}

	safeTitle := common.SanitizeFilename(exam.Title)
	fileName := fmt.Sprintf("scan_%s_%d.pdf", safeTitle, exam.ID)
	inputPDF := filepath.Join(common.GLOBAL_TEMP_SCAN, fileName)

	if _, err := os.Stat(inputPDF); err != nil {
		if os.IsNotExist(err) {
			errorLogger.Error("PDF súbor pre test neexistuje", "file_path", inputPDF, slog.String("error", err.Error()))
			return "", fmt.Errorf("PDF súbor pre test neexistuje: %s", inputPDF)
		}
		errorLogger.Error("Chyba pri kontrole PDF súboru", "file_path", inputPDF, slog.String("error", err.Error()))
		return "", fmt.Errorf("chyba pri kontrole PDF súboru: %w", err)
	}

	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
		return "", err
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
		return "", err
	}
	outputPDF := filepath.Join(absDirPath, fmt.Sprintf("student_%d_vyplnene.pdf", registrationNumber))

	var pageArgs []string
	for _, p := range pages {
		pageArgs = append(pageArgs, strconv.Itoa(p))
	}

	cmdArgs := append([]string{inputPDF, "cat"}, pageArgs...)
	cmdArgs = append(cmdArgs, "output", outputPDF)

	cmd := exec.Command("pdftk", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorLogger.Error("Chyba pri spúšťaní pdftk", "error", err.Error(), "output", string(output))
		return "", err
	}

	logger.Info("PDF slicing pomocou pdftk hotový", "output_path", outputPDF)

	// Skontroluj orientáciu prvej strany a ak je obrátená, otočíme o 180°
	if isUpsideDown(outputPDF) {
		rotated := outputPDF + "_rotated.pdf"
		rotCmd := exec.Command("pdftk", outputPDF, "rotate", "1-endsouth", "output", rotated)
		if rotOut, rotErr := rotCmd.CombinedOutput(); rotErr != nil {
			errorLogger.Error("Chyba pri rotácii PDF", "error", rotErr.Error(), "output", string(rotOut))
		} else {
			os.Remove(outputPDF)
			os.Rename(rotated, outputPDF)
			logger.Info("PDF otočené o 180°", "output_path", outputPDF)
		}
	}

	return outputPDF, nil
}

// isUpsideDown opens the first page of a PDF and checks if it is upside down
// by comparing contour counts in the upper vs lower half of the image.
func isUpsideDown(pdfPath string) bool {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return false
	}
	defer doc.Close()
	if doc.NumPage() == 0 {
		return false
	}
	img, err := doc.Image(0)
	if err != nil {
		return false
	}

	// Convert RGBA to grayscale gocv.Mat
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bytes := make([]byte, 0, w*h*3)
	for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
		for i := bounds.Min.X; i < bounds.Max.X; i++ {
			r, g, b, _ := img.At(i, j).RGBA()
			bytes = append(bytes, byte(b>>8), byte(g>>8), byte(r>>8))
		}
	}
	mat, err := gocv.NewMatFromBytes(h, w, gocv.MatTypeCV8UC3, bytes)
	if err != nil {
		return false
	}
	defer mat.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)

	// CheckUpsideDown: lower half has more contours → upside down
	upperPart := gray.Region(image.Rect(0, 0, gray.Cols(), gray.Rows()/2))
	lowerPart := gray.Region(image.Rect(0, gray.Rows()/2, gray.Cols(), gray.Rows()))
	canny := gocv.NewMat()
	defer canny.Close()

	gocv.Canny(upperPart, &canny, 100, 200)
	upperContours := gocv.FindContours(canny, gocv.RetrievalExternal, gocv.ChainApproxNone)

	gocv.Canny(lowerPart, &canny, 100, 200)
	lowerContours := gocv.FindContours(canny, gocv.RetrievalExternal, gocv.ChainApproxNone)

	return lowerContours.Size() > upperContours.Size()
}

// ExportFailedPagesToPDF extracts a subset of pages (marked as failed) from the input PDF
// and saves them into a separate output PDF file.
func ExportFailedPagesToPDF(examTitle string, examID uint, pages []int, inputPDF string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	if len(pages) == 0 {
		return nil
	}

	var pageArgs []string
	for _, p := range pages {
		pageArgs = append(pageArgs, strconv.Itoa(p+1))
	}

	cmdArgs := append([]string{inputPDF, "cat"}, pageArgs...)
	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
		return err
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
		return err
	}

	failedPagesDir := filepath.Join(absDirPath, common.FAILED_PAGES_DIR)
	if err := os.MkdirAll(failedPagesDir, 0755); err != nil {
		errorLogger.Error("Chyba pri vytváraní priečinka pre chybné strany", slog.String("error", err.Error()))
		return err
	}

	outputPDF := filepath.Join(failedPagesDir, fmt.Sprintf("%s%d_failed_pages.pdf", examTitle, examID))
	cmdArgs = append(cmdArgs, "output", outputPDF)

	cmd := exec.Command("pdftk", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorLogger.Error("Chyba pri spájaní chybných stránok", "error", err.Error(), "output", string(output))
		return err
	}

	logger.Info("Chybné strany uložené do PDF", "output", outputPDF)
	return nil
}
