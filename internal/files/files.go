package files

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/logging"
	"log/slog"
	"os"
	"strings"
)

// OpenFile reads and loads the contents of a file from the specified path.
//
// It attempts to open and read the file into memory. If successful, the file content
// is returned as a byte slice ([]byte). In case of an error, it is logged and returned.
//
// Parameters:
//   - filePath: A string representing the path to the file to be opened.
//
// Returns:
//   - []byte: The contents of the file.
//   - error: An error if the file cannot be read, otherwise nil.
//
// Notes:
//   - On failure (e.g., file not found or permission issues), a detailed error is logged.
func OpenFile(filePath string) ([]byte, error) {
	errorLogger := logging.GetErrorLogger()

	data, err := os.ReadFile(filePath)
	if err != nil {
		errorLogger.Error("Chyba pri otváraní súboru", slog.Group("CRITICAL", slog.String("error", err.Error())), slog.String("file_path", filePath))
		return nil, err
	}
	return data, nil
}

// SaveFile saves the provided byte data to the specified file path.
//
// It attempts to create or overwrite a file with the given data and standard permissions (0644).
// Success and failure are logged accordingly.
//
// Parameters:
//   - filePath: A string representing the full path where the data should be saved.
//   - data: A byte slice ([]byte) containing the data to write into the file.
//
// Returns:
//   - An error if writing to the file fails; otherwise, nil.
//
// Notes:
//   - If the file already exists, it will be overwritten.
//   - Successful and failed operations are logged for debugging and monitoring purposes.
func SaveFile(filePath string, data []byte) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	err := os.WriteFile(filePath, data, common.FILE_PERMISSION)
	if err != nil {
		errorLogger.Error("Chyba pri ukladaní súboru", slog.Group("CRITICAL", slog.String("error", err.Error())), slog.String("file_path", filePath))
		return err
	}
	logger.Info("Súbor uložený", slog.String("file_path", filePath))
	return nil
}

// DeleteFile deletes a file located at the specified file path.
//
// If the file exists, it attempts to remove it and logs the result.
// In case of a failure during deletion, it logs the error and returns it.
//
// Parameters:
//   - filePath: A string representing the path to the file that should be deleted.
//
// Returns:
//   - An error if the deletion fails; otherwise, nil.
//
// Notes:
//   - If the file does not exist, no action is taken and the function returns nil.
//   - Successful and failed operations are logged for debugging and monitoring purposes.
func DeleteFile(filePath string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	if _, err := os.Stat(filePath); err == nil {
		err = os.Remove(filePath)
		if err != nil {
			errorLogger.Error("Chyba pri mazaní súboru", slog.Group("CRITICAL", slog.String("error", err.Error())), slog.String("file_path", filePath))
			return err
		}
		logger.Info("Súbor úspešne vymazaný", slog.String("file_path", filePath))
	}
	return nil
}

func GetFilesFromConfigs() ([]string, error) {
	files, err := os.ReadDir("./configs")
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			name = strings.TrimSuffix(name, ".json")
			fileNames = append(fileNames, name)
		}
	}
	return fileNames, nil
}
