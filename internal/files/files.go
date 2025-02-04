package files

import (
	"ScanEvalApp/internal/logging"
	"log/slog"
	"os"
)

// OpenFile load file from specified path
func OpenFile(filePath string) ([]byte, error) {
	errorLogger := logging.GetErrorLogger()

	data, err := os.ReadFile(filePath)
	if err != nil {
		errorLogger.Error("Chyba pri otváraní súboru", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}
	return data, nil
}

// SaveFile saves []byte data to specified path
func SaveFile(filePath string, data []byte) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		errorLogger.Error("Chyba pri ukladaní súboru", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil
	}
	logger.Info("Súbor uložený", slog.String("file_path", filePath))
	return nil
}

// DeleteFile deletes file from specified path
func DeleteFile(filePath string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	if _, err := os.Stat(filePath); err == nil {
		// File exists, attempt to remove it
		err = os.Remove(filePath)
		if err != nil {
			errorLogger.Error("Chyba pri mazaní súboru", slog.Group("CRITICAL", slog.String("error", err.Error())))
			return err
		}
		logger.Info("Súbor úspešne vymazaný")
	}
	return nil
}
