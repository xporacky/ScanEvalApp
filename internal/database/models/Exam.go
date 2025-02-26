package models

import (
	"time"

	"gorm.io/gorm"
)

// Exam reprezentuje test
type Exam struct {
	gorm.Model
	Title         string    `gorm:"not null"`
	SchoolYear    string    `gorm:"not null"`
	Date          time.Time `gorm:"not null"`
	QuestionCount int       `gorm:"not null"`
	Questions     string    // Použitie mapy s typom char
	Students      []Student `gorm:"foreignKey:ExamID"` // Zoznam študentov
}
