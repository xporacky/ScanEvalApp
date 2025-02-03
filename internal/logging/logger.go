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

// InitLogger inicializuje logger s rôznymi úrovňami pre app.log a error.log
func InitLogger() {
	once.Do(func() {
		logsDir := "logs"

		// Vytvorenie priečinka, ak neexistuje
		if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
			log.Fatalf("CRITICAL: Nepodarilo sa vytvoriť priečinok logs: %v", err)
		}

		// Otvorenie app.log (pre DEBUG, INFO, WARNING)
		logFile, err := os.OpenFile(filepath.Join(logsDir, "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("CRITICAL: Nepodarilo sa otvoriť app.log: %v", err)
		}

		// Otvorenie error.log (pre ERROR, CRITICAL)
		errorFile, err := os.OpenFile(filepath.Join(logsDir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("CRITICAL: Nepodarilo sa otvoriť error.log: %v", err)
		}

		// Handler pre DEBUG, INFO, WARNING
		appHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Umožní logovanie od úrovne DEBUG a vyššie
		})

		// Handler pre ERROR, CRITICAL
		errorHandler := slog.NewTextHandler(errorFile, &slog.HandlerOptions{
			Level: slog.LevelError, // Umožní logovanie od úrovne ERROR a vyššie
		})

		// Inicializácia loggerov
		logger = slog.New(appHandler)
		errorLogger = slog.New(errorHandler)
	})
}

// GetLogger vráti logger pre app.log
func GetLogger() *slog.Logger {
	return logger
}

// GetErrorLogger vráti logger pre error.log
func GetErrorLogger() *slog.Logger {
	return errorLogger
}
