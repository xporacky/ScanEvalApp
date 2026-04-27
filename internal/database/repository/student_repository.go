package repository

import (
	"ScanEvalApp/internal/database/models"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"ScanEvalApp/internal/logging"
	"log/slog"
	"strings"
	"unicode"

	"gorm.io/gorm"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func CreateStudent(db *gorm.DB, student *models.Student) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Vytváranie študenta", slog.String("name", student.Name), slog.String("surname", student.Surname), "registration number", student.RegistrationNumber)
	result := db.Create(student)
	if result.Error != nil {
		errorLogger.Error("Chyba pri vytváraní študenta", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return result.Error
}

func GetStudentById(db *gorm.DB, id uint, examID uint) (*models.Student, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Hľadanie študenta", slog.Uint64("id", uint64(id)), slog.Uint64("exam ID", uint64(examID)))

	var student models.Student
	result := db.Where("ID = ? AND exam_id = ?", id, examID).First(&student)

	if result.Error != nil {
		errorLogger.Error("Študent nebol nájdený", "student id", id, slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
		return nil, result.Error
	}
	logger.Debug("Študent nájdený", slog.String("name", student.Name), slog.String("surname", student.Surname))
	return &student, nil
}

func GetStudentByRegistrationNumber(db *gorm.DB, registrationNumber uint, examID uint) (*models.Student, error) {

	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Hľadanie študenta", slog.Uint64("registration number", uint64(registrationNumber)), slog.Uint64("exam ID", uint64(examID)))

	var student models.Student

	result := db.Where("registration_number = ? AND exam_id = ?", registrationNumber, examID).First(&student)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Debug("Študent nebol nájdený", "student registration number", registrationNumber, "exam_id", examID)
			return nil, result.Error
		}
		errorLogger.Error("Študent nebol nájdený", "student registration number", registrationNumber, slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
		return nil, result.Error
	}
	logger.Debug("Študent nájdený", slog.String("name", student.Name), slog.String("surname", student.Surname))
	return &student, nil
}

func GetAllStudents(db *gorm.DB) ([]models.Student, error) {
	errorLogger := logging.GetErrorLogger()

	var students []models.Student
	result := db.Find(&students)
	if result.Error != nil {
		errorLogger.Error("Chyba pri načítavaní študentov", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return students, result.Error
}

func UpdateStudent(db *gorm.DB, student *models.Student) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Aktualizácia študenta", "registration number", student.RegistrationNumber, slog.String("name", student.Name), slog.String("surname", student.Surname), "score", student.Score)
	result := db.Save(student)
	if result.Error != nil {
		errorLogger.Error("Chyba pri aktualizácii študenta", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return result.Error
}

func DeleteStudent(db *gorm.DB, student *models.Student) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()
	logger.Debug("Mazanie študenta", "registration number", student.RegistrationNumber, slog.String("name", student.Name), slog.String("surname", student.Surname))
	result := db.Unscoped().Delete(student)
	if result.Error != nil {
		errorLogger.Error("Chyba pri mazaní študenta", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}

	return result.Error
}

func RemoveDiacritics(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	result, _, _ := transform.String(t, s)
	return strings.ToLower(result)
}

func GetStudentsQuery(db *gorm.DB, query string) ([]models.Student, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Vyhľadávanie študentov podľa dotazu", slog.String("query", query))

	var students []models.Student
	query = RemoveDiacritics(query)

	rows, err := db.Raw("SELECT * FROM students").Rows()
	if err != nil {
		errorLogger.Error("Chyba pri vyhľadávaní študentov", slog.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		db.ScanRows(rows, &student)

		if strings.Contains(RemoveDiacritics(student.Name), query) ||
			strings.Contains(RemoveDiacritics(student.Surname), query) ||
			strings.Contains(fmt.Sprintf("%d", student.RegistrationNumber), query) {
			students = append(students, student)
		}
	}
	logger.Info("Počet nájdených študentov", slog.Int("count", len(students)))

	return students, nil
}

// resolveCorrectAnswers returns the correct answer string for the given student.
// For multi-day exams (Questions is JSON), it looks up by student.Subgroup.
// If subgroup is not yet set (header not scanned), returns nil — scoring is skipped.
func resolveCorrectAnswers(exam *models.Exam, student *models.Student) []rune {
	if strings.HasPrefix(strings.TrimSpace(exam.Questions), "{") {
		var answersMap map[string]string
		if err := json.Unmarshal([]byte(exam.Questions), &answersMap); err == nil {
			if student.Subgroup == "" {
				// TODO: Subgroup not yet set — will be filled once header scanning is implemented
				return nil
			}
			if ans, ok := answersMap[student.Subgroup]; ok {
				return []rune(ans)
			}
		}
		return nil
	}
	return []rune(exam.Questions)
}

func UpdateStudentSubgroup(db *gorm.DB, studentId uint, examId uint, subgroup string) error {
	return db.Model(&models.Student{}).
		Where("id = ? AND exam_id = ?", studentId, examId).
		Update("subgroup", subgroup).Error
}

func UpdateStudentAnswers(db *gorm.DB, studentId uint, examId uint, questionNumber int, answers []rune, pageNumber int) error {
	logger := logging.GetLogger()

	student, err := GetStudentById(db, studentId, examId)
	if err != nil {
		return err
	}
	exam, err := GetExam(db, examId)
	if err != nil {
		return err
	}

	studentAnswers := []rune(student.Answers)
	startIndex := (questionNumber - len(answers)) + 1
	endIndex := questionNumber
	if startIndex < 0 || endIndex >= len(studentAnswers) {
		return fmt.Errorf("neplatný rozsah otázok pre študenta %d: start=%d end=%d answers=%d total=%d", studentId, startIndex, endIndex, len(answers), len(studentAnswers))
	}
	for i, answer := range answers {
		studentAnswers[startIndex+i] = answer
	}
	student.Answers = string(studentAnswers)

	pageNumberStr := strconv.Itoa(pageNumber)
	if student.Pages == "" {
		student.Pages = pageNumberStr
	} else {
		student.Pages += "-" + pageNumberStr
	}

	correctAnswers := resolveCorrectAnswers(exam, student)
	if correctAnswers == nil {
		logger.Info("Preskakujem skórovanie — správne odpovede nie sú dostupné", "studentID", studentId)
		UpdateStudent(db, student)
		return nil
	}

	score := 0
	for i := startIndex; i <= endIndex; i++ {
		if i >= len(correctAnswers) {
			break
		}
		if unicode.ToLower(studentAnswers[i]) == unicode.ToLower(correctAnswers[i]) {
			score++
		}
	}
	student.Score += score

	UpdateStudent(db, student)
	return nil
}

func ClearStudentForExam(db *gorm.DB, examId uint) error {
	return db.Model(&models.Student{}).
		Where("exam_id = ?", examId).
		Updates(map[string]interface{}{
			"pages": "",
			"score": 0,
		}).Error
}
