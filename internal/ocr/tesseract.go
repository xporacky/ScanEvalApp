package ocr

import (
	"ScanEvalApp/internal/logging"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
)

const PSM_SINGLE_LINE = "7"
const PSM_UNIFORM_BLOCK = "6"
const PSM_DEFAULT = "3"

func OcrImage(imagePath string, psm string) string {
	errorLogger := logging.GetErrorLogger()

	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "slk", "--psm", psm)
	out, err := cmd.Output()
	if err != nil {
		errorLogger.Error("Error during OCR process for image", slog.String("imagePath", imagePath), slog.String("error", err.Error()))
		panic(err)
	}
	return string(out)
}

func ExtractID(path string) (int, error) {
	errorLogger := logging.GetErrorLogger()
	dt := OcrImage(path, PSM_UNIFORM_BLOCK)
	re := regexp.MustCompile(`ID:\s*(\d+)`)
	match := re.FindStringSubmatch(dt)
	if len(match) < 2 {
		dt = OcrImage(path, PSM_DEFAULT)
		match = re.FindStringSubmatch(dt)
		if len(match) < 2 {
			errorLogger.Error("No ID found in the input image", slog.String("path", path))
			return 0, errors.New("no id found in the input image")
		}
	}
	var id int
	_, err := fmt.Sscan(match[1], &id)
	if err != nil {
		errorLogger.Error("Failed to convert ID to integer", slog.String("error", err.Error()))
		return 0, err
	}
	return id, nil
}

func ExtractQuestionNumber(path string) (int, error) {
	errorLogger := logging.GetErrorLogger()
	logger := logging.GetLogger()
	dt := OcrImage(path, PSM_SINGLE_LINE)
	var num int
	_, err := fmt.Sscan(dt, &num)
	if err != nil {
		errorLogger.Error("Failed to convert QuestionNumber to integer", slog.String("error", err.Error()))
		return 0, err
	}
	logger.Info("Question number", slog.Int("number", num))
	return num, nil
}
