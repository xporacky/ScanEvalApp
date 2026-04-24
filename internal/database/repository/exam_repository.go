package repository

import (
"ScanEvalApp/internal/database/models"
"ScanEvalApp/internal/logging"
"log/slog"
"time"

"gorm.io/gorm"
)

func CreateExam(db *gorm.DB, exam *models.Exam) error {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

logger.Debug("Vytváranie testu", slog.String("test", exam.Title), slog.String("cas", exam.SchoolYear))
result := db.Select("*").Create(exam)
if result.Error != nil {
errorLogger.Error("Chyba pri vytváraní testu", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
}
return result.Error
}

func GetExam(db *gorm.DB, id uint) (*models.Exam, error) {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

logger.Debug("Hľadanie testu", slog.Uint64("ID testu", uint64(id)))

var exam models.Exam
result := db.First(&exam, id)
if result.Error != nil {
errorLogger.Error("Test nebol nájdený", slog.Uint64("ID testu", uint64(id)), slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
return nil, result.Error
}
logger.Debug("Test bol nájdený", slog.String("test", exam.Title), slog.String("year", exam.SchoolYear))
return &exam, nil
}

func GetAllExams(db *gorm.DB) ([]models.Exam, error) {
errorLogger := logging.GetErrorLogger()

var exams []models.Exam
result := db.Preload("Students").Find(&exams)
if result.Error != nil {
errorLogger.Error("Chyba pri načítavaní testov", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
}
return exams, result.Error
}

func UpdateExam(db *gorm.DB, exam *models.Exam) error {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

logger.Debug("Aktualizácia testu", slog.String("exam", exam.Title), slog.String("year", exam.SchoolYear))
result := db.Save(exam)
if result.Error != nil {
errorLogger.Error("Chyba pri aktualizácii testu", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
}
return result.Error
}

func DeleteExam(db *gorm.DB, exam *models.Exam) error {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

for _, student := range exam.Students {
err := DeleteStudent(db, &student)
if err != nil {
return err
}
}

result := db.Delete(&models.Exam{}, exam.ID)
if result.Error != nil {
errorLogger.Error("Chyba pri mazaní testu", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
}

logger.Debug("Test vymazaný", slog.Uint64("test ID", uint64(exam.ID)))
return result.Error
}

func UpdateExamVersionAnswers(db *gorm.DB, examID uint, versionCode, answers string, dateTime time.Time) error {
var exam models.Exam
if err := db.First(&exam, examID).Error; err != nil {
return err
}
if err := exam.UpdateVersionAnswers(versionCode, answers, dateTime); err != nil {
return err
}
return db.Save(&exam).Error
}

func DeleteExamVersion(db *gorm.DB, examID uint, versionCode string) error {
var exam models.Exam
if err := db.First(&exam, examID).Error; err != nil {
return err
}
if err := exam.DeleteVersion(versionCode); err != nil {
return err
}
return db.Save(&exam).Error
}

func GetExamVersions(db *gorm.DB, examID uint) (map[string]models.VersionData, error) {
var exam models.Exam
if err := db.First(&exam, examID).Error; err != nil {
return nil, err
}
return exam.GetVersionsMap()
}


