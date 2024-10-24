package main

import "time"

// Student reprezentuje študenta
type Student struct {
	ID                 uint      `gorm:"primaryKey"`
	Name               string    `gorm:"not null"`
	Surname            string    `gorm:"not null"`
	BirthDate          time.Time `gorm:"not null"`
	RegistrationNumber string    `gorm:"not null;unique"`
	Room               string    `gorm:"not null"`
	Score              int
	Answers            map[int]rune
}
