package csv

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/logging"

	"gorm.io/gorm"
)

// ImportStudentsFromCSV parses student records from the given CSV content
// and stores them in the database.
func ImportStudentsFromCSV(db *gorm.DB, csvContent string, examID uint) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	reader := csv.NewReader(strings.NewReader(csvContent))
	rows, err := reader.ReadAll()
	if err != nil {
		errorLogger.Error("Chyba pri čítaní CSV súboru", slog.String("error", err.Error()))
		return err
	}

	for i, row := range rows {
		if i == 0 {
			continue // Preskoc hlavicku csv
		}
		birthDate, err := time.Parse("2006-01-02", row[2])
		if err != nil {
			errorLogger.Error("Chyba pri parsovaní dátumu narodenia", slog.String("error", err.Error()))
			return err
		}
		registrationNumber, err := strconv.Atoi(row[3])
		if err != nil {
			errorLogger.Error("Chyba pri parsovaní registračného čísla", slog.String("error", err.Error()))
			return err
		}

		student := models.Student{
			Name:               row[0],
			Surname:            row[1],
			BirthDate:          birthDate,
			RegistrationNumber: registrationNumber,
			Room:               row[4],
			ExamID:             examID,
		}

		if err := repository.CreateStudent(db, &student); err != nil {
			errorLogger.Error("Chyba pri ukladaní študenta", slog.String("studentName", student.Name), slog.String("error", err.Error()))
			return err
		}
	}

	logger.Info("Import študentov z CSV dokončený", slog.Int("studentCount", len(rows)-1))

	return nil
}

// ExportStudentsToCSV exports all students associated with the given exam
// into a CSV file.
func ExportStudentsToCSV(db *gorm.DB, exam models.Exam) (string, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	var students []models.Student
	err := db.Where("exam_id = ?", exam.ID).Find(&students).Error
	if err != nil {
		errorLogger.Error("Chyba pri načítaní študentov", slog.String("error", err.Error()))
		return "", err
	}

	safeTitle := strings.ReplaceAll(exam.Title, " ", "_")

	fileName := fmt.Sprintf("%s%s_ID%d.csv", common.GLOBAL_EXPORT_DIR, safeTitle, exam.ID)

	file, err := os.Create(fileName)
	if err != nil {
		errorLogger.Error("Chyba pri vytváraní CSV súboru", slog.String("fileName", fileName), slog.String("error", err.Error()))
		return "", err
	}
	defer file.Close()

	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Hlavicka CSV
	err = writer.Write([]string{"ID", "Meno", "Priezvisko", "Registračné číslo", "Skóre"})

	if err != nil {
		errorLogger.Error("Chyba pri zápise hlavičky CSV", slog.String("error", err.Error()))
		return "", err
	}

	for _, student := range students {
		record := []string{
			strconv.Itoa(int(student.ID)),
			student.Name,
			student.Surname,
			strconv.Itoa(student.RegistrationNumber),
			strconv.Itoa(student.Score),
		}
		err := writer.Write(record)

		if err != nil {
			errorLogger.Error("Chyba pri zápise záznamu študenta do CSV", slog.String("studentName", student.Name), slog.String("error", err.Error()))
			return "", err
		}
	}

	logger.Info("Export študentov do CSV úspešný", slog.String("fileName", fileName), slog.Int("studentCount", len(students)))

	return fileName, nil
}
