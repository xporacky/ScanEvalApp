package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/config"
	"ScanEvalApp/internal/database/migrations"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/files/csv"
	"ScanEvalApp/internal/files/pdf"
	"ScanEvalApp/internal/latex"
	"ScanEvalApp/internal/logging"
	"ScanEvalApp/internal/scanprocessing"
	"ScanEvalApp/internal/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx   context.Context
	db    *gorm.DB
	dbErr error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logging.InitLogger()
	a.db, a.dbErr = migrations.MigrateDB()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

type ExamSummary struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	SchoolYear    string    `json:"schoolYear"`
	Date          time.Time `json:"date"`
	QuestionCount int       `json:"questionCount"`
	OptionCount   int       `json:"optionCount"`
	StudentCount  int       `json:"studentCount"`
	IsMultiDay    bool      `json:"isMultiDay"`
}

type ExamTemplate struct {
	Title             string   `json:"title"`
	SchoolYear        string   `json:"schoolYear"`
	DateTime          string   `json:"dateTime"`
	QuestionCount     int      `json:"questionCount"`
	OptionCount       int      `json:"optionCount"`
	ShowName          bool     `json:"showName"`
	Answers           []string `json:"answers"`
	StudentCSVContent string   `json:"studentCSVContent"`
}

func (a *App) ListExams() ([]ExamSummary, error) {
	if a.dbErr != nil {
		return nil, a.dbErr
	}
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	exams, err := repository.GetAllExams(a.db)
	if err != nil {
		return nil, err
	}

	summaries := make([]ExamSummary, 0, len(exams))
	for _, exam := range exams {
		summaries = append(summaries, ExamSummary{
			ID:            exam.ID,
			Title:         exam.Title,
			SchoolYear:    exam.SchoolYear,
			Date:          exam.Date,
			QuestionCount: exam.QuestionCount,
			OptionCount:   exam.OptionCount,
			StudentCount:  len(exam.Students),
			IsMultiDay:    exam.IsMultiDay,
		})
	}
	return summaries, nil
}

type StudentSummary struct {
	ID                 uint      `json:"id"`
	Name               string    `json:"name"`
	Surname            string    `json:"surname"`
	BirthDate          time.Time `json:"birthDate"`
	Room               string    `json:"room"`
	RegistrationNumber int       `json:"registrationNumber"`
	ExamID             uint      `json:"examId"`
	Score              int       `json:"score"`
	Pages              string    `json:"pages"`
}

func (a *App) ListStudents() ([]StudentSummary, error) {
	if a.dbErr != nil {
		return nil, a.dbErr
	}
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	students, err := repository.GetAllStudents(a.db)
	if err != nil {
		return nil, err
	}

	summaries := make([]StudentSummary, 0, len(students))
	for _, student := range students {
		summaries = append(summaries, StudentSummary{
			ID:                 student.ID,
			Name:               student.Name,
			Surname:            student.Surname,
			BirthDate:          student.BirthDate,
			Room:               student.Room,
			RegistrationNumber: student.RegistrationNumber,
			ExamID:             student.ExamID,
			Score:              student.Score,
			Pages:              student.Pages,
		})
	}
	return summaries, nil
}

func (a *App) CreateExam(title, schoolYear, dateISO string, questionCount, optionCount int) error {
	if a.dbErr != nil {
		return a.dbErr
	}
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if title == "" || schoolYear == "" || questionCount <= 0 {
		return fmt.Errorf("invalid input")
	}

	date, err := time.Parse("2006-01-02", dateISO)
	if err != nil {
		return err
	}

	exam := &models.Exam{
		Title:         title,
		SchoolYear:    schoolYear,
		Date:          date,
		QuestionCount: questionCount,
		OptionCount:   optionCount,
	}

	return repository.CreateExam(a.db, exam)
}

func (a *App) CreateExamWithCSV(title, schoolYear, dateTime string, questionCount, optionCount int, answers []string, csvContent string, showName bool) error {
	if a.dbErr != nil {
		return a.dbErr
	}
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if title == "" || schoolYear == "" {
		return fmt.Errorf("invalid input")
	}

	return services.CreateExamWithCSV(a.db, title, schoolYear, dateTime, questionCount, optionCount, answers, csvContent, showName)
}

func (a *App) CreateMultiDayExamWithCSV(title, schoolYear string, questionCount, optionCount int, showName bool, csvContent string, subgroupAnswers map[string]string) error {
	if a.dbErr != nil {
		return a.dbErr
	}
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if title == "" || schoolYear == "" {
		return fmt.Errorf("invalid input")
	}
	return services.CreateMultiDayExamWithCSV(a.db, title, schoolYear, questionCount, optionCount, showName, csvContent, subgroupAnswers)
}

func (a *App) UpdateMultiDayAnswers(examID uint, subgroupAnswers map[string]string) error {
	if a.dbErr != nil {
		return a.dbErr
	}
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if examID == 0 {
		return fmt.Errorf("invalid exam id")
	}
	return services.UpdateMultiDayAnswers(a.db, examID, subgroupAnswers)
}

func (a *App) PrintMultiDayExamPDFs(examID uint) ([]string, error) {
	if a.dbErr != nil {
		return nil, a.dbErr
	}
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return latex.GenerateMultiDayPDFs(a.db, examID)
}

func (a *App) DeleteExam(examID uint) error {
	if a.dbErr != nil {
		return a.dbErr
	}
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	var exam models.Exam
	if err := a.db.Preload("Students").First(&exam, examID).Error; err != nil {
		return err
	}

	return repository.DeleteExam(a.db, &exam)
}

func (a *App) PrintExamPDF(examID uint) (string, error) {
	if a.dbErr != nil {
		return "", a.dbErr
	}
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	return latex.ParallelGeneratePDFs(a.db, examID)
}

func (a *App) OpenPath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}
	return exec.Command("xdg-open", path).Start()
}

type ExamStats struct {
	ExamID       uint    `json:"examId"`
	StudentCount int     `json:"studentCount"`
	AvgScore     float64 `json:"avgScore"`
	MinScore     int     `json:"minScore"`
	MaxScore     int     `json:"maxScore"`
}

func (a *App) GetExamStats(examID uint) (ExamStats, error) {
	if a.dbErr != nil {
		return ExamStats{}, a.dbErr
	}
	if a.db == nil {
		return ExamStats{}, fmt.Errorf("database not initialized")
	}

	var students []models.Student
	if err := a.db.Where("exam_id = ?", examID).Find(&students).Error; err != nil {
		return ExamStats{}, err
	}

	stats := ExamStats{ExamID: examID, StudentCount: len(students)}
	if len(students) == 0 {
		return stats, nil
	}

	minScore := students[0].Score
	maxScore := students[0].Score
	sum := 0
	for _, s := range students {
		sum += s.Score
		if s.Score < minScore {
			minScore = s.Score
		}
		if s.Score > maxScore {
			maxScore = s.Score
		}
	}
	stats.MinScore = minScore
	stats.MaxScore = maxScore
	stats.AvgScore = float64(sum) / float64(len(students))
	return stats, nil
}

func (a *App) GenerateExamStatisticsPDF(examID uint, selected []string) (string, error) {
	if a.dbErr != nil {
		return "", a.dbErr
	}
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}
	if examID == 0 {
		return "", fmt.Errorf("invalid exam id")
	}

	var exam models.Exam
	if err := a.db.Preload("Students").First(&exam, examID).Error; err != nil {
		return "", err
	}

	return latex.GenerateStatistics(selected, &exam)
}

func (a *App) GetExamAnswers(examID uint) (string, error) {
	if a.dbErr != nil {
		return "", a.dbErr
	}
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	exam, err := repository.GetExam(a.db, examID)
	if err != nil {
		return "", err
	}

	return exam.Questions, nil
}

func (a *App) UpdateExamAnswers(examID uint, answers []string) error {
	if a.dbErr != nil {
		return a.dbErr
	}
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if examID == 0 {
		return fmt.Errorf("invalid exam id")
	}
	return services.UpdateExamAnswers(a.db, examID, answers)
}

func (a *App) PrintStudentSheet(studentID uint) (string, error) {
	if a.dbErr != nil {
		return "", a.dbErr
	}
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}
	if studentID == 0 {
		return "", fmt.Errorf("invalid student id")
	}

	return latex.PrintSheet(a.db, studentID)
}

func (a *App) DownloadStudentSheet(studentID uint) (string, error) {
	if a.dbErr != nil {
		return "", a.dbErr
	}
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}
	if studentID == 0 {
		return "", fmt.Errorf("invalid student id")
	}

	return pdf.SlicePdfForStudent(a.db, studentID)
}

func (a *App) ExportMultiDayResultsCSV(examID uint) (string, error) {
	if a.dbErr != nil {
		return "", a.dbErr
	}
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}
	return csv.ExportMultiDayResultsCSV(a.db, examID)
}

func (a *App) ExportExamStudentsCSV(examID uint) (string, error) {
	if a.dbErr != nil {
		return "", a.dbErr
	}
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
	}
	if examID == 0 {
		return "", fmt.Errorf("invalid exam id")
	}

	exam, err := repository.GetExam(a.db, examID)
	if err != nil {
		return "", err
	}

	return csv.ExportStudentsToCSV(a.db, *exam)
}

func (a *App) ExportExamTemplateCSV(title, schoolYear, dateTime string, questionCount, optionCount int, answers []string, studentCSVContent string, showName bool) (string, error) {
	template := csv.ExamTemplate{
		Title:             title,
		SchoolYear:        schoolYear,
		DateTime:          dateTime,
		QuestionCount:     questionCount,
		OptionCount:       optionCount,
		ShowName:          showName,
		Answers:           answers,
		StudentCSVContent: studentCSVContent,
	}

	return csv.ExportExamTemplateCSV(template)
}

func (a *App) ParseExamTemplateCSV(csvContent string) (ExamTemplate, error) {
	template, err := csv.ParseExamTemplateCSV(csvContent)
	if err != nil {
		return ExamTemplate{}, err
	}

	return ExamTemplate{
		Title:             template.Title,
		SchoolYear:        template.SchoolYear,
		DateTime:          template.DateTime,
		QuestionCount:     template.QuestionCount,
		OptionCount:       template.OptionCount,
		ShowName:          template.ShowName,
		Answers:           template.Answers,
		StudentCSVContent: template.StudentCSVContent,
	}, nil
}

func (a *App) ListConfigs() ([]string, error) {
	return files.GetFilesFromConfigs()
}

func (a *App) EvaluateExam(examID uint, pdfPath, configName string) error {
	if a.dbErr != nil {
		return a.dbErr
	}
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if examID == 0 || pdfPath == "" || configName == "" {
		return fmt.Errorf("missing input")
	}

	go func() {
		runtime.EventsEmit(a.ctx, "evaluation_progress", "Spustam vyhodnotenie...")

		if err := scanprocessing.LoadConfig(configName); err != nil {
			runtime.EventsEmit(a.ctx, "evaluation_error", err.Error())
			return
		}

		exam, err := repository.GetExam(a.db, examID)
		if err != nil {
			runtime.EventsEmit(a.ctx, "evaluation_error", err.Error())
			return
		}

		progressChan := make(chan string, 100)
		var counter int
		hadFailures := false

		go func() {
			for msg := range progressChan {
				runtime.EventsEmit(a.ctx, "evaluation_progress", msg)
			}
		}()

		scanprocessing.ProcessPDF(pdfPath, exam, a.db, progressChan, &counter, &hadFailures)
		close(progressChan)

		failedPath := ""
		if hadFailures {
			if dirPath, err := config.LoadLastPath(); err == nil {
				if absDirPath, err := filepath.Abs(dirPath); err == nil {
					safeTitle := common.SanitizeFilename(exam.Title)
					failedPath = filepath.Join(absDirPath, fmt.Sprintf("%s%d_failed_pages.pdf", safeTitle, exam.ID))
				}
			}
		}

		runtime.EventsEmit(a.ctx, "evaluation_done", map[string]interface{}{
			"examId":      exam.ID,
			"hadFailures": hadFailures,
			"failedPath":  failedPath,
		})
	}()

	return nil
}

func (a *App) PickPDF() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("runtime not ready")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Vyber PDF",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF", Pattern: "*.pdf"},
		},
	})
}

func (a *App) PickFolder() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("runtime not ready")
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Vyber priecinok pre ukladanie",
	})
}

func (a *App) GetSavePath() (string, error) {
	return config.LoadLastPath()
}

func (a *App) SetSavePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}
	return config.SaveLastPath(path)
}

func (a *App) PrintLegend() (string, error) {
	legendContent, err := os.ReadFile("assets/latex/legend.tex")
	if err != nil {
		return "", fmt.Errorf("failed to read legend.tex: %w", err)
	}

	pdfBytes, err := latex.CompileLatexToPDF(legendContent)
	if err != nil {
		return "", fmt.Errorf("failed to compile legend: %w", err)
	}

	dirPath, cfgErr := config.LoadLastPath()
	if cfgErr != nil || dirPath == "" {
		dirPath = os.TempDir()
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		absDirPath = os.TempDir()
	}

	outputPath := filepath.Join(absDirPath, "legenda.pdf")
	if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to save legend PDF: %w", err)
	}

	return outputPath, nil
}
