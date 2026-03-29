package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/files/csv"

	"gorm.io/gorm"
)

func CreateExamWithCSV(db *gorm.DB, title, schoolYear, dateTime string, questionCount, optionCount int, answers []string, csvContent string, showName bool) error {
	if !isValidSchoolYear(schoolYear) {
		return fmt.Errorf("invalid school year")
	}
	parsedDateTime, ok := parseDateTime(dateTime)
	if !ok {
		return fmt.Errorf("invalid date")
	}
	if questionCount <= 0 {
		return fmt.Errorf("invalid question count")
	}

	answersStr := strings.Join(answers, "")
	exam := models.Exam{
		Title:         title,
		SchoolYear:    schoolYear,
		Date:          parsedDateTime,
		QuestionCount: questionCount,
		Questions:     answersStr,
		OptionCount:   optionCount,
		ShowName:      showName,
	}

	if err := repository.CreateExam(db, &exam); err != nil {
		return err
	}

	return csv.ImportStudentsFromCSV(db, csvContent, exam.ID)
}

func UpdateExamAnswers(db *gorm.DB, examID uint, answers []string) error {
	answersStr := strings.Join(answers, "")
	return db.Model(&models.Exam{}).Where("id = ?", examID).Update("questions", answersStr).Error
}

func isValidSchoolYear(s string) bool {
	re := regexp.MustCompile(`^\d{4}/\d{2}$`)
	return re.MatchString(s)
}

func parseDateTime(dateTime string) (time.Time, bool) {
	// Skús formát len s dátumom (dd.MM.yyyy)
	if t, err := time.Parse("02.01.2006", dateTime); err == nil {
		return t, true
	}
	// Fallback: starý formát s časom (dd.MM.yyyy HH:mm)
	if t, err := time.Parse("02.01.2006 15:04", dateTime); err == nil {
		return t, true
	}
	return time.Time{}, false
}
