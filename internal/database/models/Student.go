package models

import (
	"time"

	"gorm.io/gorm"
)

// Student reprezentuje študenta
type Student struct {
	gorm.Model
	Name               string    `gorm:"not null"`
	Surname            string    `gorm:"not null"`
	BirthDate          time.Time `gorm:"not null"`
	RegistrationNumber int       `gorm:"not null;uniqueIndex:idx_students_exam_registration"`
	Room               string    `gorm:"not null"`
	Score              int
	Answers            string
	ExamID             uint `gorm:"not null;uniqueIndex:idx_students_exam_registration"`
	Pages              string
	ExamDate           time.Time `gorm:"default:null"`
	ExamTime           string
	Subgroup           string
}
