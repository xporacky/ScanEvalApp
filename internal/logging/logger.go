package logging

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
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

		// Rotujúce log súbory pre app.log
		appLogWriter := &lumberjack.Logger{
			Filename:   filepath.Join(logsDir, "app.log"),
			MaxSize:    5,  // Max 5 MB
			MaxBackups: 3,  // Udržiava max 3 staré logy
			MaxAge:     7,  // Ukladá logy max 7 dní
			Compress:   true, // Kompresia starých logov
		}

		// Rotujúce log súbory pre error.log
		errorLogWriter := &lumberjack.Logger{
			Filename:   filepath.Join(logsDir, "error.log"),
			MaxSize:    5,  // Max 5 MB
			MaxBackups: 3,  // Udržiava max 3 staré logy
			MaxAge:     7,  // Ukladá logy max 7 dní
			Compress:   true, // Kompresia starých logov
		}

		// Handler pre DEBUG, INFO, WARNING
		appHandler := slog.NewTextHandler(appLogWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug, // Umožní logovanie od úrovne DEBUG a vyššie
		})

		// Handler pre ERROR, CRITICAL
		errorHandler := slog.NewTextHandler(errorLogWriter, &slog.HandlerOptions{
			Level: slog.LevelError, // Umožní logovanie od úrovne ERROR a vyššie
		})

		// Inicializácia loggerov
		logger = slog.New(appHandler)
		errorLogger = slog.New(errorHandler)
	})
}

// GetLogger vráti logger pre app.log
func GetLogger() *slog.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}

// GetErrorLogger vráti logger pre error.log
func GetErrorLogger() *slog.Logger {
	if errorLogger == nil {
		InitLogger()
	}
	return errorLogger
}
