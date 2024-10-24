package main

import "gorm.io/gorm"

// Test reprezentuje test
type Test struct {
	gorm.Model
	Title         string    `gorm:"not null"`
	SchoolYear    string    `gorm:"not null"`
	QuestionCount int       `gorm:"not null"`
	Questions     string    // Použitie mapy s typom char
	Students      []Student `gorm:"foreignKey:TestID"` // Zoznam študentov
}
