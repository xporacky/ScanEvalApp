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
	RegistrationNumber string    `gorm:"not null;unique"`
	Room               string    `gorm:"not null"`
	Score              int
	Answers            string
	TestID             uint `gorm:"not null"`
}
