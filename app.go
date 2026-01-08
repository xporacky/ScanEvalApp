package main

import (
	//"context"
	"fmt"
)

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetAppInfo returns basic app information
func (a *App) GetAppInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":    "ScanEvalApp Pro",
		"version": "1.0.0",
		"ready":   true,
	}
}

// TestDatabaseConnection tests if database is working
func (a *App) TestDatabaseConnection() (bool, error) {
	if a.db == nil {
		return false, fmt.Errorf("database not initialized")
	}
	
	// Simple test query
	var result int
	if err := a.db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return false, err
	}
	
	return true, nil
}
