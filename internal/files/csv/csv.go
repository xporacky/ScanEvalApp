package csv

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/config"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/logging"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ExamTemplate struct {
	Title             string
	SchoolYear        string
	DateTime          string
	QuestionCount     int
	OptionCount       int
	ShowName          bool
	Answers           []string
	StudentCSVContent string
}

// ImportStudentsFromCSV parses student records from the given CSV content
// and stores them in the database.
func ImportStudentsFromCSV(db *gorm.DB, csvContent string, examID uint) error {
	exam, err := repository.GetExam(db, examID)
	if err != nil {
		return fmt.Errorf("chyba pri načítaní testu: %w", err)
	}

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
			Answers:            strings.Repeat("0", exam.QuestionCount),
		}
		student.Answers = strings.Repeat("0", exam.QuestionCount)
		_, err = repository.GetStudentByRegistrationNumber(db, uint(registrationNumber), examID)
		if err == nil {
			logger.Info("Študent už je v teste, preskakujem import duplicitného záznamu",
				slog.Int("registrationNumber", registrationNumber),
				slog.Uint64("examID", uint64(examID)))
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			errorLogger.Error("Chyba pri overení existencie študenta",
				slog.Int("registrationNumber", registrationNumber),
				slog.String("error", err.Error()))
			return err
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
	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
		return "", err
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
		return "", err
	}

	safeTitle := common.SanitizeFilename(exam.Title)

	fileName := filepath.Join(absDirPath, fmt.Sprintf("%s_ID%d.csv", safeTitle, exam.ID))

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

func ExportExamTemplateCSV(template ExamTemplate) (string, error) {
	errorLogger := logging.GetErrorLogger()

	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
		return "", err
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
		return "", err
	}

	fileName := "test_sablona.csv"
	if strings.TrimSpace(template.Title) != "" {
		fileName = common.SanitizeFilename(template.Title) + "_sablona.csv"
	}
	filePath := filepath.Join(absDirPath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		errorLogger.Error("Chyba pri vytváraní CSV šablóny", slog.String("fileName", filePath), slog.String("error", err.Error()))
		return "", err
	}
	defer file.Close()

	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	defer writer.Flush()

	rows := [][]string{
		{"section", "key", "value"},
		{"meta", "title", template.Title},
		{"meta", "school_year", template.SchoolYear},
		{"meta", "date_time", template.DateTime},
		{"meta", "question_count", strconv.Itoa(template.QuestionCount)},
		{"meta", "option_count", strconv.Itoa(template.OptionCount)},
		{"meta", "show_name", strconv.FormatBool(template.ShowName)},
	}

	for index, answer := range template.Answers {
		rows = append(rows, []string{"answer", strconv.Itoa(index + 1), strings.TrimSpace(answer)})
	}

	rows = append(rows, []string{"payload", "students_csv", template.StudentCSVContent})

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			errorLogger.Error("Chyba pri zápise CSV šablóny", slog.String("fileName", filePath), slog.String("error", err.Error()))
			return "", err
		}
	}

	return filePath, nil
}

func ParseExamTemplateCSV(csvContent string) (ExamTemplate, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))
	rows, err := reader.ReadAll()
	if err != nil {
		return ExamTemplate{}, err
	}
	if len(rows) == 0 {
		return ExamTemplate{}, fmt.Errorf("csv sablona je prazdna")
	}

	template := ExamTemplate{
		OptionCount: 5,
		ShowName:    true,
	}

	for i, row := range rows {
		if len(row) < 3 {
			continue
		}
		section := strings.TrimSpace(strings.TrimPrefix(row[0], "\ufeff"))
		key := strings.TrimSpace(row[1])
		value := row[2]

		if i == 0 && strings.EqualFold(section, "section") {
			continue
		}

		switch section {
		case "meta":
			switch key {
			case "title":
				template.Title = value
			case "school_year":
				template.SchoolYear = value
			case "date_time":
				template.DateTime = value
			case "question_count":
				template.QuestionCount, err = strconv.Atoi(strings.TrimSpace(value))
				if err != nil {
					return ExamTemplate{}, fmt.Errorf("neplatny question_count")
				}
			case "option_count":
				template.OptionCount, err = strconv.Atoi(strings.TrimSpace(value))
				if err != nil {
					return ExamTemplate{}, fmt.Errorf("neplatny option_count")
				}
			case "show_name":
				template.ShowName, err = strconv.ParseBool(strings.TrimSpace(value))
				if err != nil {
					return ExamTemplate{}, fmt.Errorf("neplatny show_name")
				}
			}
		case "answer":
			template.Answers = append(template.Answers, strings.TrimSpace(value))
		case "payload":
			if key == "students_csv" {
				template.StudentCSVContent = value
			}
		}
	}

	if template.QuestionCount <= 0 {
		template.QuestionCount = len(template.Answers)
	}
	if template.QuestionCount <= 0 {
		return ExamTemplate{}, fmt.Errorf("sablona neobsahuje pocet otazok")
	}

	for len(template.Answers) < template.QuestionCount {
		template.Answers = append(template.Answers, "")
	}
	if len(template.Answers) > template.QuestionCount {
		template.Answers = template.Answers[:template.QuestionCount]
	}

	if template.OptionCount < 2 {
		template.OptionCount = 2
	}
	if template.OptionCount > 8 {
		template.OptionCount = 8
	}

	return template, nil
}
