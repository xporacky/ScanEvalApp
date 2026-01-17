package main

import (
	"ScanEvalApp/internal/database/migrations"
	window "ScanEvalApp/internal/gui"
	"ScanEvalApp/internal/logging"
	"log/slog"

	"gioui.org/app"
)

func main() {
	logging.InitLogger()
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Info("---------------------------------------------------")
	errorLogger.Error("---------------------------------------------------")

	logger.Info("Aplikácia spustená")

	// inicializacia a migracia db
	logger.Info("Spúšťam migráciu databázy.")
	db, err := migrations.MigrateDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		panic("failed to connect to database")
	}
	logger.Info("Migrácia databázy dokončená.")

	logger.Info("Spúšťam GUI.")
	go window.RunWindow(db)
	app.Main()
}
