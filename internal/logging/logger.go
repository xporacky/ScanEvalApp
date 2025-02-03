package logging

import (
	"os"
	"log"
	"log/slog"
	"path/filepath"
	"sync"
)

var (
	logger      *slog.Logger
	errorLogger *slog.Logger
	once        sync.Once
)

// InitLogger inicializuje logovanie len raz
func InitLogger() {
	once.Do(func() {
		logsDir := "logs"

		// Vytvorenie priečinka, ak neexistuje
		if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
			log.Fatalf("CRITICAL: Nepodarilo sa vytvoriť priečinok logs: %v", err)
		}

		// Otvorenie app.log
		logFile, err := os.OpenFile(filepath.Join(logsDir, "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("CRITICAL: Nepodarilo sa otvoriť app.log: %v", err)
		}

		// Otvorenie error.log
		errorFile, err := os.OpenFile(filepath.Join(logsDir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("CRITICAL: Nepodarilo sa otvoriť error.log: %v", err)
		}

		// Inicializácia logerov
		logger = slog.New(slog.NewTextHandler(logFile, nil))
		errorLogger = slog.New(slog.NewTextHandler(errorFile, nil))
	})
}

// GetLogger vráti hlavný logger
func GetLogger() *slog.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}

// GetErrorLogger vráti logger pre chyby
func GetErrorLogger() *slog.Logger {
	if errorLogger == nil {
		InitLogger()
	}
	return errorLogger
}
