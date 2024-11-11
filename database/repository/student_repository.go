package repository

import (
	"gorm.io/gorm"
	"ScanEvalApp/database/models"
)

func CreateStudent(db *gorm.DB, student *models.Student) error {
	result := db.Create(student)
	return result.Error
}

func GetStudent(db *gorm.DB, id uint) (*models.Student, error) {
	var student models.Student
	result := db.First(&student, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &student, nil
}

func GetAllStudents(db *gorm.DB) ([]models.Student, error) {
	var students []models.Student
	result := db.Find(&students)
	return students, result.Error
}

func UpdateStudent(db *gorm.DB, student *models.Student) error {
	result := db.Save(student)
	return result.Error
}

func DeleteStudent(db *gorm.DB, id uint) error {
	result := db.Delete(&models.Student{}, id)
	return result.Error
}