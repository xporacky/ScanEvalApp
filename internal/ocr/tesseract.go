package ocr

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/logging"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// PSM_SINGLE_LINE specifies Tesseract's Page Segmentation Mode 7,
// treating the image as a single line of text.
const PSM_SINGLE_LINE = "7"

// PSM_UNIFORM_BLOCK specifies Tesseract's Page Segmentation Mode 6,
// treating the image as a single uniform block of text.
const PSM_UNIFORM_BLOCK = "6"

// PSM_DEFAULT specifies Tesseract's default Page Segmentation Mode 3,
// fully automatic page segmentation but no orientation and script detection.
const PSM_DEFAULT = "3"

// OcrImage performs OCR (Optical Character Recognition) on an image file using Tesseract.
//
// It runs the Tesseract command-line tool with the specified page segmentation mode (PSM)
// and returns the extracted text. If any error occurs (e.g., resolving the path or executing
// the Tesseract command), it logs the error and returns it.
//
// Parameters:
//   - imagePath: A string path to the image file to be processed.
//   - psm: A string representing the Page Segmentation Mode for Tesseract (e.g., "3", "6").
//
// Returns:
//   - string: The text extracted from the image.
//   - error: An error if the OCR process fails or the image path cannot be resolved.
//
// Notes:
//   - The OCR language is set to Slovak ("slk").
func OcrImage(imagePath string, psm string) (string, error) {
	errorLogger := logging.GetErrorLogger()

	imagePath, err := filepath.Abs(imagePath)
	if err != nil {
		errorLogger.Error("Chyba pri získavaní absolútnej cesty k obrázku", slog.String("error", err.Error()))
		return "", err
	}
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "slk", "--psm", psm)
	out, err := cmd.Output()
	if err != nil {
		errorLogger.Error("Error during OCR process for image", slog.String("imagePath", imagePath), slog.String("error", err.Error()))
		return "", err
	}
	return string(out), nil
}

func ocrImageDigitsOnly(imagePath string, psm string) (string, error) {
	errorLogger := logging.GetErrorLogger()

	imagePath, err := filepath.Abs(imagePath)
	if err != nil {
		errorLogger.Error("Chyba pri získavaní absolútnej cesty k obrázku", slog.String("error", err.Error()))
		return "", err
	}
	cmd := exec.Command("tesseract", imagePath, "stdout", "--psm", psm,
		"-c", "tessedit_char_whitelist=0123456789/")
	out, err := cmd.Output()
	if err != nil {
		errorLogger.Error("Error during OCR process for image", slog.String("imagePath", imagePath), slog.String("error", err.Error()))
		return "", err
	}
	return string(out), nil
}

// ExtractID extracts a numeric ID from an image by performing OCR.
//
// It first attempts OCR using a uniform block page segmentation mode to locate a pattern "ID: <number>".
// If the first OCR attempt fails to find the ID, it retries with a default OCR configuration.
// The extracted ID is parsed into an integer. If no valid ID is found or conversion fails, an error is returned.
//
// Parameters:
//   - path: A string representing the path to the image file containing the ID.
//
// Returns:
//   - int: The extracted ID number.
//   - error: An error if OCR fails, no ID is found, or parsing the ID fails.
//
// Notes:
//   - Logs detailed errors for troubleshooting if extraction or parsing fails.
// parseIDSlashFormat extracts a registration number from whitelist-only OCR output.
// With tessedit_char_whitelist=0123456789/, Tesseract maps box-enclosed digits correctly
// and we can search directly for the dddd/ddd pattern.
func parseIDSlashFormat(text string) (int, bool) {
	re := regexp.MustCompile(`(\d{4}/\d{3})`)
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return 0, false
	}
	digitsOnly := strings.ReplaceAll(match[1], "/", "")
	var id int
	_, err := fmt.Sscan(digitsOnly, &id)
	if err != nil {
		return 0, false
	}
	return id, true
}

// parseIDFromOCRText extracts the registration number from noisy OCR output.
// Fallback for when whitelist OCR fails — looks for "ID:" label and strips non-digits.
func parseIDFromOCRText(text string) (int, bool) {
	re := regexp.MustCompile(`ID:\s*([\d\s/]+)`)
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return 0, false
	}
	digitsOnly := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, match[1])
	if digitsOnly == "" {
		return 0, false
	}
	var id int
	_, err := fmt.Sscan(digitsOnly, &id)
	if err != nil {
		return 0, false
	}
	return id, true
}

func ExtractID(path string) (int, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Primary: whitelist-only OCR forces digit mapping, then match dddd/ddd pattern.
	dt, err := ocrImageDigitsOnly(path, PSM_UNIFORM_BLOCK)
	if err != nil {
		return 0, err
	}
	logger.Info("OCR digits-only raw text (PSM_UNIFORM_BLOCK)", slog.String("text", dt))
	if id, ok := parseIDSlashFormat(dt); ok {
		return id, nil
	}

	dt, err = ocrImageDigitsOnly(path, PSM_DEFAULT)
	if err != nil {
		return 0, err
	}
	logger.Info("OCR digits-only raw text (PSM_DEFAULT)", slog.String("text", dt))
	if id, ok := parseIDSlashFormat(dt); ok {
		return id, nil
	}

	// Fallback: full OCR with "ID:" label search.
	dt, err = OcrImage(path, PSM_UNIFORM_BLOCK)
	if err != nil {
		return 0, err
	}
	logger.Info("OCR raw text (PSM_UNIFORM_BLOCK)", slog.String("text", dt))
	if id, ok := parseIDFromOCRText(dt); ok {
		return id, nil
	}

	dt, err = OcrImage(path, PSM_DEFAULT)
	if err != nil {
		return 0, err
	}
	logger.Info("OCR raw text (PSM_DEFAULT)", slog.String("text", dt))
	if id, ok := parseIDFromOCRText(dt); ok {
		return id, nil
	}

	errorLogger.Error("No ID found in the input image", slog.String("path", path))
	return 0, errors.New("no id found in the input image")
}

// ExtractQuestionNumber performs OCR on the specified image path to extract and parse a question number.
//
// It uses OCR (configured for single-line text) to read the content of the image,
// then attempts to convert the recognized text into an integer representing the question number.
// If OCR fails or parsing the number fails, the error is logged and returned.
//
// Parameters:
//   - path: A string representing the path to the image file containing the question number.
//
// Returns:
//   - int: The extracted question number.
//   - error: An error if OCR or conversion to integer fails.
//
// Notes:
//   - Logs successful extraction or any errors encountered during the process.
func ExtractQuestionNumber(path string) (int, error) {
	errorLogger := logging.GetErrorLogger()
	logger := logging.GetLogger()
	dt, err := OcrImage(path, PSM_SINGLE_LINE)
	if err != nil {
		return common.QUESTION_NUMBER_NOT_FOUND, err
	}
	var num int
	_, err = fmt.Sscan(dt, &num)
	if err != nil {
		errorLogger.Error("Failed to convert QuestionNumber to integer", slog.String("error", err.Error()))
		return common.QUESTION_NUMBER_NOT_FOUND, err
	}
	logger.Info("Question number", slog.Int("number", num))
	return num, nil
}
