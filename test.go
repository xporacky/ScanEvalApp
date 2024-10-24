package main

// Test reprezentuje test
type Test struct {
	ID          uint              `gorm:"primaryKey"`
	Title       string            `gorm:"not null"`
	SchoolYear  string            `gorm:"not null"`
	QuestionCount int              `gorm:"not null"`
	Questions   map[int]rune      // Použitie mapy s typom char
	Students    []Student         // Zoznam študentov
}
