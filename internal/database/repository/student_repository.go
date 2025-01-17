package repository

import (
	"ScanEvalApp/internal/database/models"

	"gorm.io/gorm"
)

func CreateStudent(db *gorm.DB, student *models.Student) error {
	result := db.Create(student)
	return result.Error
}

func GetStudent(db *gorm.DB, registrationNumber uint, testID uint) (*models.Student, error) {
	var student models.Student
	result := db.Where("registration_number = ? AND test_id = ?", registrationNumber, testID).First(&student)
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
