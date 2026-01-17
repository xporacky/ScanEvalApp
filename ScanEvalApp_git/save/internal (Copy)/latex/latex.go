package latex

import (
"bytes"
"fmt"
"html/template"
"io"
"os"
"os/exec"
"path/filepath"
"strings"
"sync"
"time"
"unicode"

"golang.org/x/text/unicode/norm"
"gorm.io/gorm"

"ScanEvalApp/internal/common"
"ScanEvalApp/internal/config"
"ScanEvalApp/internal/database/models"
"ScanEvalApp/internal/files"
"ScanEvalApp/internal/logging"
)

// TemplateData represents the data used for replacing placeholders in a LaTeX template.
type TemplateData struct {
ID        string
Meno      string
Datum     string
Miestnost string
Cas       string
Bloky     int
QrCode    string
}

// removeDiacritics removes diacritics from the input string and replaces spaces with underscores.
func removeDiacritics(input string) string {
t := norm.NFD.String(input)
t = strings.Map(func(r rune) rune {
if unicode.IsMark(r) {
return -1
}
return r
}, t)
// Replace spaces with underscores
t = strings.ReplaceAll(t, " ", "_")
return t
}

// FindStudentByRegistrationNumber finds a student in the database by their registration number.
func FindStudentByRegistrationNumber(db *gorm.DB, registrationNumber int) (*models.Student, error) {
errorLogger := logging.GetErrorLogger()
var student models.Student
// Query the database for the student with the specified registration number
if err := db.Where("registration_number = ?", registrationNumber).First(&student).Error; err != nil {
// Log an error if the student is not found
errorLogger.Error("Student not found with ", "registration_number", registrationNumber, "error", err.Error())
return nil, err
}
return &student, nil
}

// CompileLatexToPDF compiles a LaTeX template into a PDF.
func CompileLatexToPDF(latexContent []byte) ([]byte, error) {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

// Create a temporary file to store the LaTeX content
texFile, err := os.CreateTemp(common.TEMPORARY_PDF_PATH, "*.tex")
if err != nil {
errorLogger.Error("Failed to create temporary LaTeX file", "error", err.Error())
return nil, err
}

defer func() {
if err := files.DeleteFile(texFile.Name()); err != nil {
errorLogger.Error("Error deleting temporary LaTeX file", "error", err.Error())
}
}()

// Write LaTeX content into the file
if _, err = texFile.Write(latexContent); err != nil {
errorLogger.Error("Error writing to .tex file", "error", err.Error())
return nil, err
}

// Set the output directory for the PDF file
if err := texFile.Close(); err != nil {
errorLogger.Error("Error closing temporary LaTeX file", "error", err.Error())
return nil, err
}

logger.Info("LaTeX file created", "file_path", texFile.Name())

outputDir, err := os.MkdirTemp(common.TEMPORARY_PDF_PATH, "latex_output")
if err != nil {
errorLogger.Error("Failed to create temporary output directory", "error", err.Error())
return nil, err
}
defer os.RemoveAll(outputDir)

// DEBUG: Save the LaTeX content for manual inspection
debugTexPath := filepath.Join(outputDir, "debug_latex.tex")
if err := os.WriteFile(debugTexPath, latexContent, 0644); err == nil {
logger.Debug("LaTeX content saved for debugging", "debug_file", debugTexPath)
}

// Run pdflatex command to compile the LaTeX file - CAPTURE BOTH STDOUT AND STDERR
cmd := exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error", "-output-directory", outputDir, texFile.Name())

// Capture both stdout and stderr
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

logger.Debug("Running pdflatex command", "command", cmd.String())

if err = cmd.Run(); err != nil {
latexStdout := stdout.String()
latexStderr := stderr.String()

errorLogger.Error("=== LaTeX COMPILATION FAILED ===",
"error", err.Error(),
"stdout", latexStdout,
"stderr", latexStderr,
"tex_file", texFile.Name(),
"output_dir", outputDir)

// Check if output files were created
pdfPath := filepath.Join(outputDir, filepath.Base(texFile.Name())[:len(filepath.Base(texFile.Name()))-4]+".pdf")
if _, pdfErr := os.Stat(pdfPath); pdfErr == nil {
errorLogger.Error("PDF file was created despite error", "pdf_path", pdfPath)
} else {
errorLogger.Error("PDF file was NOT created", "expected_pdf_path", pdfPath)
}

// List files in output directory
if files, err := os.ReadDir(outputDir); err == nil {
errorLogger.Error("Files in output directory:")
for _, file := range files {
errorLogger.Error("Output file:", "file", file.Name(), "is_dir", file.IsDir())
}
}

return nil, fmt.Errorf("latex compilation failed: %s", latexStderr)
}

pdfPath := filepath.Join(outputDir, filepath.Base(texFile.Name())[:len(filepath.Base(texFile.Name()))-4]+".pdf")
pdfBytes, err := os.ReadFile(pdfPath)
if err != nil {
errorLogger.Error("Error reading compiled PDF",
"error", err.Error(),
"pdf_path", pdfPath)

// List files to see what was generated
if files, listErr := os.ReadDir(outputDir); listErr == nil {
errorLogger.Error("Files in output directory after compilation:")
for _, file := range files {
errorLogger.Error("Generated file:", "file", file.Name(), "is_dir", file.IsDir())
}
}
return nil, err
}

logger.Info("PDF compiled successfully", "pdf_size_bytes", len(pdfBytes))
return pdfBytes, nil
}

// ReplaceTemplatePlaceholders replaces placeholders in a LaTeX template with the provided data.
func ReplaceTemplatePlaceholders(templateContent []byte, data TemplateData) ([]byte, error) {
errorLogger := logging.GetErrorLogger()

// Parse the LaTeX template content
tmpl, err := template.New("latex").Parse(string(templateContent))
if err != nil {
errorLogger.Error("Error parsing LaTeX template", "error", err.Error())
return nil, err
}

var output bytes.Buffer
// Apply the data to the template
err = tmpl.Execute(&output, data)
if err != nil {
errorLogger.Error("Error replacing placeholders in template", "error", err.Error())
return nil, err
}

return output.Bytes(), nil
}

// MergePDFs merges two PDFs into a single PDF using the pdfunite utility.
func MergePDFs(pdf1Path, pdf2Path, outputPath string) error {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

// Execute the pdfunite command to merge PDFs
cmd := exec.Command("pdfunite", pdf1Path, pdf2Path, outputPath)
cmd.Stdout = nil
cmd.Stderr = nil
logger.Debug("Merging PDFs", "pdf1", pdf1Path, "pdf2", pdf2Path)

if err := cmd.Run(); err != nil {
errorLogger.Error("Error merging PDFs", "error", err.Error())
return err
}
logger.Info("PDFs merged", "output_path", outputPath)
return nil
}

// ParallelGeneratePDFs generates PDFs for students in parallel by processing each student in separate goroutines.
func ParallelGeneratePDFs(db *gorm.DB, examID uint) (string, error) {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

// DEBUG: Check current working directory
if wd, err := os.Getwd(); err == nil {
logger.Debug("Current working directory", "working_directory", wd)
} else {
errorLogger.Error("Error getting working directory", "error", err.Error())
}

logger.Debug("Starting PDF generation", "exam_id", examID)

// Load exam details including option count
var exam models.Exam
if err := db.First(&exam, examID).Error; err != nil {
errorLogger.Error("Error fetching exam details", "exam_id", examID, "error", err.Error())
return "", err
}

// COMPREHENSIVE DEBUGGING - Check exam data
logger.Debug("Exam details loaded",
"exam_id", exam.ID,
"exam_title", exam.Title,
"option_count", exam.OptionCount,
"question_count", exam.QuestionCount)

// Check if OptionCount is valid
if exam.OptionCount < common.MIN_OPTION_COUNT || exam.OptionCount > common.MAX_OPTION_COUNT {
errorLogger.Error("Invalid option count in exam",
"option_count", exam.OptionCount,
"min_allowed", common.MIN_OPTION_COUNT,
"max_allowed", common.MAX_OPTION_COUNT)
return "", fmt.Errorf("invalid option count: %d", exam.OptionCount)
}

// Determine template path based on option count
templatePath := common.GetTemplatePath(exam.OptionCount)

// COMPREHENSIVE PATH DEBUGGING
logger.Debug("=== TEMPLATE PATH DEBUGGING ===")
logger.Debug("Template path result",
"template_path", templatePath,
"input_option_count", exam.OptionCount)

// Check if template path is empty
if templatePath == "" {
errorLogger.Error("Template path is EMPTY - this is the problem!")
return "", fmt.Errorf("template path is empty")
}

// Check if template file exists with detailed info
fileInfo, err := os.Stat(templatePath)
if os.IsNotExist(err) {
errorLogger.Error("=== TEMPLATE FILE DOES NOT EXIST ===",
"template_path", templatePath,
"option_count", exam.OptionCount,
"error", err.Error())

// List available templates for debugging
templateDir := filepath.Dir(templatePath)
logger.Debug("Checking template directory", "directory", templateDir)

// Check if directory exists
if dirInfo, dirErr := os.Stat(templateDir); os.IsNotExist(dirErr) {
errorLogger.Error("=== TEMPLATE DIRECTORY DOES NOT EXIST ===",
"directory", templateDir,
"error", dirErr.Error())

// Show what directories do exist
parentDir := filepath.Dir(templateDir)
logger.Debug("Checking parent directory", "parent_directory", parentDir)
if parentFiles, parentErr := os.ReadDir(parentDir); parentErr == nil {
logger.Debug("Files in parent directory:")
for _, file := range parentFiles {
logger.Debug("Parent dir file:", "file", file.Name(), "is_dir", file.IsDir())
}
}
} else if dirErr != nil {
errorLogger.Error("Error checking template directory", "error", dirErr.Error())
} else {
logger.Debug("Template directory exists",
"directory", templateDir,
"is_dir", dirInfo.IsDir())
}

// List files in template directory
files, err := os.ReadDir(templateDir)
if err != nil {
errorLogger.Error("Error reading template directory", "error", err.Error())
} else {
logger.Debug("=== AVAILABLE FILES IN TEMPLATE DIRECTORY ===")
for _, file := range files {
logger.Debug("Found file:",
"file", file.Name(),
"is_dir", file.IsDir(),
"full_path", filepath.Join(templateDir, file.Name()))
}

// Check if any template files exist with similar names
logger.Debug("=== LOOKING FOR SIMILAR TEMPLATE FILES ===")
for i := common.MIN_OPTION_COUNT; i <= common.MAX_OPTION_COUNT; i++ {
expectedFile := fmt.Sprintf("template_%d.tex", i)
fullExpectedPath := filepath.Join(templateDir, expectedFile)
if _, err := os.Stat(fullExpectedPath); err == nil {
logger.Debug("Template file EXISTS",
"file", expectedFile,
"full_path", fullExpectedPath)
} else {
logger.Debug("Template file NOT FOUND",
"file", expectedFile,
"full_path", fullExpectedPath)
}
}
}
return "", fmt.Errorf("template file does not exist: %s", templatePath)
} else if err != nil {
errorLogger.Error("Error checking template file",
"template_path", templatePath,
"error", err.Error())
return "", err
} else {
logger.Info("=== TEMPLATE FILE FOUND ===",
"template_path", templatePath,
"file_size", fileInfo.Size(),
"file_mod_time", fileInfo.ModTime())
}

logger.Info("Using template", "template_path", templatePath, "option_count", exam.OptionCount)

// Continue with the rest of your function...
var rooms []string
// Fetch unique room names from the database
if err := db.Model(&models.Student{}).Where("exam_id = ?", examID).Distinct().Pluck("room", &rooms).Error; err != nil {
errorLogger.Error("Error fetching distinct rooms", "error", err.Error())
return "", err
}

// Synchronize goroutines using a WaitGroup
var wg sync.WaitGroup
var pdfMergeMutex sync.Mutex
var processedCount int64

// Variable to store the path of the main PDF
var mainPDFPath string
var mainPDFSet bool

logger.Debug("Starting parallel PDF generation",
"option_count", exam.OptionCount,
"template_path", templatePath,
"num_rooms", len(rooms))

// Measure the total time for PDF generation and merging
startTime := time.Now()

// Loop through each room and process students due to correct ordering of PDFs
for _, room := range rooms {
logger.Info("Processing students in room", "room", room)

// Fetch all students in the current room
var students []models.Student
if err := db.Where("room = ? AND exam_id = ?", room, examID).Find(&students).Error; err != nil {
errorLogger.Error("Error fetching students", "error", err.Error())
return "", err
}

logger.Debug("Found students for room", "room", room, "num_students", len(students))

// Generate PDFs concurrently for each student
for _, student := range students {
wg.Add(1)
go func(student models.Student) {
defer wg.Done()

studentStartTime := time.Now()

// Load LaTeX template with dynamic path based on option count
logger.Debug("Reading template file for student",
"student_id", student.ID,
"template_path", templatePath)

latexTemplate, err := os.ReadFile(templatePath)
if err != nil {
errorLogger.Error("Error reading LaTeX template for student",
"student_id", student.ID,
"error", err.Error(),
"template_path", templatePath)
return
}

logger.Debug("Successfully read template",
"student_id", student.ID,
"template_size_bytes", len(latexTemplate))

// Prepare the data to replace placeholders in the LaTeX template
data := TemplateData{
ID:        fmt.Sprintf("%d", student.RegistrationNumber),
Meno:      fmt.Sprintf("%s %s", student.Name, student.Surname),
Datum:     exam.Date.Format("02. 01. 2006"),
Miestnost: student.Room,
Cas:       exam.Date.Format("15:04"),
Bloky:     exam.QuestionCount,
QrCode:    fmt.Sprintf("%d", student.ID),
}

// Replace placeholders in the LaTeX template with the student data
updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
if err != nil {
errorLogger.Error("Error replacing placeholders for student", "student_id", student.ID, "error", err.Error())
return
}

// DEBUG: Log template replacement results
logger.Debug("Template replacement completed",
"student_id", student.ID,
"original_template_size", len(latexTemplate),
"updated_latex_size", len(updatedLatex))

// Log a sample of the generated LaTeX to see if placeholders were replaced
if len(updatedLatex) > 200 {
logger.Debug("Generated LaTeX sample (first 200 chars):",
"student_id", student.ID,
"sample", string(updatedLatex[:200]))
} else {
logger.Debug("Generated LaTeX content:",
"student_id", student.ID,
"content", string(updatedLatex))
}

// Compile the LaTeX template into a PDF
studentPDF, err := CompileLatexToPDF(updatedLatex)
if err != nil {
errorLogger.Error("Error generating PDF for student", "student_id", student.ID, "error", err.Error())
return
}

// Save the generated PDF for the student
studentPDFPath := filepath.Join(common.TEMPORARY_PDF_PATH, fmt.Sprintf("student_%d.pdf", student.ID))
if err := os.WriteFile(studentPDFPath, studentPDF, common.FILE_PERMISSION); err != nil {
errorLogger.Error("Error saving PDF for student", "student_id", student.ID, "error", err.Error())
return
}

// Lock the PDF merge operation to ensure it happens sequentially
pdfMergeMutex.Lock()
defer pdfMergeMutex.Unlock()

// Set the first student PDF as the main PDF
if !mainPDFSet {
mainPDFPath = studentPDFPath
mainPDFSet = true
logger.Info("Set initial main PDF for student", "student_id", student.ID)
} else {
// Merge the new student PDF with the existing main PDF
mergedPDFPath := filepath.Join(common.TEMPORARY_PDF_PATH, "merged.pdf")
if err := MergePDFs(mainPDFPath, studentPDFPath, mergedPDFPath); err != nil {
errorLogger.Error("Error merging PDF for student", "student_id", student.ID, "error", err.Error())
return
}

// Rename the merged PDF to be the main PDF
if err := os.Rename(mergedPDFPath, mainPDFPath); err != nil {
errorLogger.Error("Error updating main PDF for student", "student_id", student.ID, "error", err.Error())
return
}

// Delete the temp student pdf
defer func() {
if err := files.DeleteFile(studentPDFPath); err != nil {
errorLogger.Error("Error removing temporary PDF for student", "student_id", student.ID, "error", err.Error())
}
}()

}

// Increment the processed PDF count and log it
processedCount++
studentDuration := time.Since(studentStartTime)
logger.Debug("Generovanie PDF",
"spracovaných", processedCount,
"celkovo", len(students),
"exam", exam.Title,
"id študenta", student.ID,
"dokončené za", studentDuration)
}(student)
}

// Waiting for all goroutines
wg.Wait()
}

// Create a final pdf
dirPath, err := config.LoadLastPath()
if err != nil {
errorLogger.Error("Chyba načítania configu", "error", err.Error())
return "", err
}

absDirPath, err := filepath.Abs(dirPath)
if err != nil {
errorLogger.Error("Chyba pri konverzii cesty", "error", err.Error())
return "", err
}

safeTitle := common.SanitizeFilename(exam.Title)
finalPDFPath := filepath.Join(absDirPath, fmt.Sprintf("%s%d.pdf", safeTitle, exam.ID))

srcFile, err := os.Open(mainPDFPath)
if err != nil {
errorLogger.Error("error opening source PDF", "error", err.Error())
return "", err
}
defer srcFile.Close()

dstFile, err := os.Create(finalPDFPath)
if err != nil {
errorLogger.Error("error creating destination PDF", "error", err.Error())
return "", err
}
defer dstFile.Close()

if _, err := io.Copy(dstFile, srcFile); err != nil {
errorLogger.Error("error copying PDF", "error", err.Error())
return "", err
}

// Files will be closed automatically via defer statements
if err := os.Remove(mainPDFPath); err != nil {
errorLogger.Error("error removing original PDF", "error", err.Error())
return "", err
}
// Measure the total time
duration := time.Since(startTime)
logger.Debug("Celkový čas generovania PDF", "duration", duration)

absFinalPDFPath, err := filepath.Abs(finalPDFPath)
if err != nil {
errorLogger.Error("Nepodarilo sa získať absolútnu cestu k výslednému PDF", "error", err.Error())
return "", err
}
logger.Info("Výsledné PDF úspešne uložené do", "output_PDF_path", absFinalPDFPath)

return absFinalPDFPath, nil
}

// PrintSheet generates a PDF for a student based on their registration number.
func PrintSheet(db *gorm.DB, registrationNumber int) (string, error) {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

// Find student by registrationNumber
student, err := FindStudentByRegistrationNumber(db, registrationNumber)
if err != nil {
errorLogger.Error("Error finding student", "registration_number", registrationNumber, "error", err.Error())
return "", err
}

logger.Info("Student found", "student_id", student.ID, "registration_number", student.RegistrationNumber)

// Load the exam for student from the database
var exam models.Exam
if err := db.First(&exam, student.ExamID).Error; err != nil {
errorLogger.Error("Error fetching exam for student", "student_id", student.ID, "error", err.Error())
return "", err
}
logger.Info("Exam fetched for student", "exam_id", exam.ID, "exam_title", exam.Title)

// Determine template path based on exam's option count
templatePath := common.GetTemplatePath(exam.OptionCount)

// Load the latex template with dynamic path
latexTemplate, err := os.ReadFile(templatePath)
if err != nil {
errorLogger.Error("Error reading LaTeX template for student", "student_id", student.ID, "error", err.Error())
return "", err
}

logger.Info("LaTeX template loaded", "template_path", templatePath, "option_count", exam.OptionCount)

// Create data template based on student's data from the database
data := TemplateData{
ID:        fmt.Sprintf("%d", student.RegistrationNumber),
Meno:      fmt.Sprintf("%s %s", student.Name, student.Surname),
Datum:     exam.Date.Format("02. 01. 2006"),
Miestnost: student.Room,
Cas:       exam.Date.Format("15:04"),
Bloky:     exam.QuestionCount,
QrCode:    fmt.Sprintf("%d", student.ID),
}

logger.Info("Template data prepared", "student_id", student.ID, "registration_number", student.RegistrationNumber)

// Replace placeholder by student's data
updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
if err != nil {
errorLogger.Error("Error replacing placeholders for student", "student_id", student.ID, "error", err.Error())
return "", err
}
logger.Info("Placeholders replaced successfully for student", "student_id", student.ID)

// Compile the pdf
studentPDF, err := CompileLatexToPDF(updatedLatex)
if err != nil {
errorLogger.Error("Error generating PDF for student", "student_id", student.ID, "error", err.Error())
return "", err
}
logger.Info("PDF generated for student", "student_id", student.ID)

dirPath, err := config.LoadLastPath()
if err != nil {
errorLogger.Error("Chyba načítania configu", "error", err.Error())
return "", err
}

absDirPath, err := filepath.Abs(dirPath)
if err != nil {
errorLogger.Error("Chyba pri konverzii cesty", "error", err.Error())
return "", err
}

// Save the student's compiled PDF
studentPDFPath := filepath.Join(absDirPath, fmt.Sprintf("student_%d.pdf", student.RegistrationNumber))
err = os.WriteFile(studentPDFPath, studentPDF, common.FILE_PERMISSION)
if err != nil {
errorLogger.Error("Chyba pri ukladaní PDF pre študenta",
"student_id", student.ID,
"error", err.Error())
return "", err
}

logger.Info("PDF úspešne uložené pre študenta",
"student_id", student.ID,
"pdf_path", studentPDFPath,
)

return studentPDFPath, nil
}
