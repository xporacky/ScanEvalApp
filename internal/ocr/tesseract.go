package ocr

import (
	"ScanEvalApp/internal/logging"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"regexp"
)

const PSM_SINGLE_LINE = "7"
const PSM_UNIFORM_BLOCK = "6"
const PSM_DEFAULT = "3"

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

func ExtractID(path string) (int, error) {
	errorLogger := logging.GetErrorLogger()
	dt, err := OcrImage(path, PSM_UNIFORM_BLOCK)
	if err != nil {
		return 0, err
	}
	re := regexp.MustCompile(`ID:\s*(\d+)`)
	match := re.FindStringSubmatch(dt)
	if len(match) < 2 {
		dt, err := OcrImage(path, PSM_DEFAULT)
		if err != nil {
			return 0, err
		}
		match = re.FindStringSubmatch(dt)
		if len(match) < 2 {
			errorLogger.Error("No ID found in the input image", slog.String("path", path))
			return 0, errors.New("no id found in the input image")
		}
	}
	var id int
	_, err = fmt.Sscan(match[1], &id)
	if err != nil {
		errorLogger.Error("Failed to convert ID to integer", slog.String("error", err.Error()))
		return 0, err
	}
	return id, nil
}

func ExtractQuestionNumber(path string) (int, error) {
	errorLogger := logging.GetErrorLogger()
	logger := logging.GetLogger()
	dt, err := OcrImage(path, PSM_SINGLE_LINE)
	if err != nil {
		return 0, err
	}
	var num int
	_, err = fmt.Sscan(dt, &num)
	if err != nil {
		errorLogger.Error("Failed to convert QuestionNumber to integer", slog.String("error", err.Error()))
		return 0, err
	}
	logger.Info("Question number", slog.Int("number", num))
	return num, nil
}
