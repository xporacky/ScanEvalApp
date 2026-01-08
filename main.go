package main

import (
	"ScanEvalApp/internal/database/migrations"
	"ScanEvalApp/internal/logging"
	"context"
	"embed"
	"log/slog"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"gorm.io/gorm"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx context.Context
	db  *gorm.DB
}

// NewApp creates a new App application struct
func NewApp(db *gorm.DB) *App {
	return &App{db: db}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logging.GetLogger().Info("Wails aplikácia inicializovaná")
}

func (a *App) shutdown(ctx context.Context) {
	logging.GetLogger().Info("Aplikácia sa vypína")
	if sqlDB, err := a.db.DB(); err == nil {
		sqlDB.Close()
	}
}

func main() {
	logging.InitLogger()
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Info("---------------------------------------------------")
	errorLogger.Error("---------------------------------------------------")

	logger.Info("ScanEvalApp Pro - Wails verzia")

	// Initialize database
	logger.Info("Spúšťam migráciu databázy.")
	db, err := migrations.MigrateDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", 
			slog.Group("CRITICAL", slog.String("error", err.Error())))
		panic("failed to connect to database")
	}
	logger.Info("Migrácia databázy dokončená.")

	// Create application
	app := NewApp(db)

	// Run Wails app
	err = wails.Run(&options.App{
		Title:  "ScanEvalApp Pro",
		Width:  1400,
		Height: 900,
		MinWidth:  1024,
		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
	})

	if err != nil {
		errorLogger.Error("Chyba pri spustení aplikácie", slog.String("error", err.Error()))
		panic(err)
	}
}
