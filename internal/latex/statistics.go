package latex

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"
	"ScanEvalApp/internal/config"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"path/filepath"
)

// GenerateStatistics generates statistics based on selected options
// and compiles a LaTeX report for the exam data.
// It returns the file path of the generated PDF or an error.
func GenerateStatistics(selectedStats []string, exam *models.Exam) (string, error) {
	errorLogger := logging.GetErrorLogger()

	students := exam.Students

	scores := getScores(students)
	correctAnswers := strings.TrimSpace(exam.Questions)
	correctAnswers = strings.ToLower(correctAnswers)

	statsData := make(map[string]interface{})

	// Initialize statistics options
	statsData["includeMax"] = false
	statsData["includeMin"] = false
	statsData["includeAvg"] = false
	statsData["includeMedian"] = false
	statsData["includeScoreDistribution"] = false
	statsData["includePerQuestionDistribution"] = false
	statsData["includeOverallSuccess"] = false
	statsData["includePerQuestionSuccess"] = false

	// Collect requested statistics
	for _, stat := range selectedStats {
		switch stat {
		case "Maximum bodov":
			max := calculateMax(scores)
			statsData["includeMax"] = true
			statsData["max"] = max

		case "Minimum bodov":
			min := calculateMin(scores)
			statsData["includeMin"] = true
			statsData["min"] = min

		case "Priemer":
			avg := calculateAverage(scores)
			statsData["includeAvg"] = true
			statsData["avg"] = avg

		case "Medián":
			median := calculateMedian(scores)
			statsData["includeMedian"] = true
			statsData["median"] = median

		case "Graf rozdelenia bodov celkovo":
			statsData["includeScoreDistribution"] = true
			statsData["scores"] = scores

		case "Graf rozdelenia za jednotlivé príklady":
			successPerQuestion := calculateSuccessPerQuestion(students, correctAnswers, exam.QuestionCount)
			statsData["includePerQuestionDistribution"] = true
			statsData["successPerQuestion"] = successPerQuestion

		case "Úspešnosť absolútna aj relatívna":
			absolute, relative := calculateOverallSuccess(students, exam.QuestionCount)
			statsData["includeOverallSuccess"] = true
			statsData["sumPoints"] = exam.QuestionCount * len(students)
			statsData["absoluteSuccess"] = absolute
			statsData["relativeSuccess"] = relative * 100

		case "Úspešnosť absolútna aj relatívna pre jednotlivé príklady":
			absolutePerQuestion, relativePerQuestion := calculatePerQuestionSuccess(students, correctAnswers, exam.QuestionCount)
			statsData["includePerQuestionSuccess"] = true
			statsData["absolutePerQuestion"] = absolutePerQuestion
			statsData["relativePerQuestion"] = relativePerQuestion

		default:
			errorLogger.Error("Neznáma štatistika", slog.String("stat", stat))
		}
	}

	// Generate LaTeX report
	latexContent, err := GenerateLatexReport(exam, statsData)
	if err != nil {
		errorLogger.Error("Chyba pri generovaní LaTeXu", slog.String("error", err.Error()))
		return "", err
	}

	// Compile LaTeX to PDF
	pdfBytes, err := CompileLatexToPDF(latexContent)
	if err != nil {
		errorLogger.Error("Chyba pri kompilácii LaTeXu", slog.String("error", err.Error()))
		return "", err
	}

	// Save PDF to file
	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
		return "", err
	}
	outputPath := filepath.Join(dirPath, fmt.Sprintf("stats_%s.pdf", exam.Title))
	if err := os.WriteFile(outputPath, pdfBytes, FILE_PERMISSION); err != nil {
		errorLogger.Error("Chyba pri ukladaní PDF", slog.String("path", outputPath), slog.String("error", err.Error()))
		return "", err
	}

	return outputPath, nil
}

// getScores extracts the scores from a slice of students and returns them as a list of integers.
func getScores(students []models.Student) []int {
	scores := make([]int, len(students))
	for i, s := range students {
		scores[i] = s.Score
	}
	return scores
}

// calculateMax calculates and returns the maximum score from the list of scores.
func calculateMax(scores []int) int {
	if len(scores) == 0 {
		return 0
	}
	max := scores[0]
	for _, s := range scores {
		if s > max {
			max = s
		}
	}
	return max
}

// calculateMin calculates and returns the minimum score from the list of scores.
func calculateMin(scores []int) int {
	if len(scores) == 0 {
		return 0
	}
	min := scores[0]
	for _, s := range scores {
		if s < min {
			min = s
		}
	}
	return min
}

// calculateAverage calculates and returns the average score from the list of scores.
func calculateAverage(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}
	sum := 0
	for _, s := range scores {
		sum += s
	}
	return float64(sum) / float64(len(scores))
}

// calculateMedian calculates and returns the median score from the list of scores.
func calculateMedian(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}
	sort.Ints(scores)
	mid := len(scores) / 2
	if len(scores)%2 == 0 {
		return float64(scores[mid-1]+scores[mid]) / 2
	}
	return float64(scores[mid])
}

// calculateOverallSuccess calculates and returns the total number of correct answers and the relative success rate.
func calculateOverallSuccess(students []models.Student, totalQuestions int) (int, float64) {
	totalPossible := totalQuestions * len(students)
	totalCorrect := 0
	for _, s := range students {
		totalCorrect += s.Score
	}
	relative := float64(totalCorrect) / float64(totalPossible)
	return totalCorrect, relative
}

// latexEscape escapes special characters for LaTeX compatibility.
func latexEscape(str string) string {
	replacer := strings.NewReplacer(
		"\\", "\\textbackslash{}",
		"&", "\\&",
		"%", "\\%",
		"$", "\\$",
		"#", "\\#",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"~", "\\textasciitilde{}",
		"^", "\\textasciicircum{}",
	)
	return replacer.Replace(str)
}

// GenerateLatexReport generates the LaTeX content for the exam report.
func GenerateLatexReport(exam *models.Exam, statsData map[string]interface{}) ([]byte, error) {
	var builder strings.Builder

	builder.WriteString(`
	\documentclass{article}
	\usepackage[utf8]{inputenc}
	\usepackage{tabularx}
	\usepackage{pgfplots}
	\usepackage{graphicx}
	\pgfplotsset{compat=1.18}
	\title{Štatistiky testu: ` + latexEscape(exam.Title) + `}
	\date{}
	\begin{document}
	\maketitle
	`)

	// Basic statistics table
	if statsData["includeMax"].(bool) || statsData["includeMin"].(bool) || statsData["includeAvg"].(bool) || statsData["includeMedian"].(bool) || statsData["includeOverallSuccess"].(bool) {
		builder.WriteString(`\section{Základné štatistiky}`)
	}
	builder.WriteString(`
	\begin{tabular}{|l|r|}
	\hline
	`)

	// Include maximum, minimum, average, and median values if requested
	if statsData["includeMax"].(bool) {
		builder.WriteString(fmt.Sprintf("Maximum bodov & %d \\\\\n\\hline\n", statsData["max"].(int)))
	}
	if statsData["includeMin"].(bool) {
		builder.WriteString(fmt.Sprintf("Minimum bodov & %d \\\\\n\\hline\n", statsData["min"].(int)))
	}
	if statsData["includeAvg"].(bool) {
		builder.WriteString(fmt.Sprintf("Priemer & %.2f \\\\\n\\hline\n", statsData["avg"].(float64)))
	}
	if statsData["includeMedian"].(bool) {
		builder.WriteString(fmt.Sprintf("Medián & %.2f \\\\\n\\hline\n", statsData["median"].(float64)))
	}
	if statsData["includeOverallSuccess"].(bool) {
		builder.WriteString(fmt.Sprintf("Celkový počet bodov & %d \\\\\n\\hline\n", statsData["sumPoints"].(int)))
		builder.WriteString(fmt.Sprintf("Absolútna úspešnosť & %d \\\\\n\\hline\n", statsData["absoluteSuccess"].(int)))
		builder.WriteString(fmt.Sprintf("Relatívna úspešnosť & %.2f \\\\\n\\hline\n", statsData["relativeSuccess"].(float64)))
	}

	builder.WriteString(`\end{tabular}`)

	// Score distribution graph
	if statsData["includeScoreDistribution"].(bool) {
		labels, coords := buildPlotData(statsData["scores"].([]int))

		builder.WriteString(`
		\section{Rozdelenie bodov}
		\begin{tikzpicture}
		\begin{axis}[
			ybar,
			xlabel={Rozsahy bodov},
			ylabel={Počet študentov},
			width=\textwidth,
			height=8cm,
			bar width=0.8cm,
			xtick=data,
			xticklabels={` + labels + `},
			nodes near coords,
		]
		\addplot coordinates {
		` + coords + `
		};
		\end{axis}
		\end{tikzpicture}
		`)
	}

	// Success rate per question graph
	if statsData["includePerQuestionDistribution"].(bool) {
		successRates := statsData["successPerQuestion"].([]float64)
		labels, coords := buildPerQuestionPlotData(successRates)

		builder.WriteString(`
		\section{Úspešnosť za jednotlivé príklady}
		\begin{tikzpicture}
		\begin{axis}[
			xbar,
			xlabel={Úspešnosť (\%)},
			ylabel={Príklad},
			width=\textwidth,
			height=` + fmt.Sprintf("%d", len(successRates)/2) + `cm,
			bar width=0.4cm,
			ytick={1,...,` + fmt.Sprintf("%d", len(successRates)) + `},
			yticklabels={` + labels + `},
			yticklabel style={font=\footnotesize, align=right},
			xmin=0, xmax=100,
			nodes near coords,
			nodes near coords align={horizontal},
			enlarge y limits=0.02,
		]
		\addplot coordinates {
		` + coords + `
		};
		\end{axis}
		\end{tikzpicture}
		`)
	}

	if statsData["includePerQuestionSuccess"].(bool) {
		absolute := statsData["absolutePerQuestion"].([]int)
		relative := statsData["relativePerQuestion"].([]float64)

		// First table: original order of exam questions
		builder.WriteString(`
		\section{Úspešnosť za jednotlivé príklady (pôvodné poradie)}
		\begin{tabular}{|l|r|r|}
		\hline
		\textbf{Príklad} & \textbf{Absolútna} & \textbf{Relatívna (\%)} \\ \hline
		`)
		for q := 0; q < len(absolute); q++ {
			builder.WriteString(fmt.Sprintf("%d & %d & %.2f \\\\\n\\hline\n", q+1, absolute[q], relative[q]))
		}
		builder.WriteString(`\end{tabular}`)

		// Second table: exam questions ordered by relative success
		type question struct {
			number   int
			absolute int
			relative float64
		}
		var questions []question
		for q := range absolute {
			questions = append(questions, question{
				number:   q + 1,
				absolute: absolute[q],
				relative: relative[q],
			})
		}
		sort.Slice(questions, func(i, j int) bool {
			return questions[i].relative > questions[j].relative
		})

		builder.WriteString(`
		\section{Úspešnosť za jednotlivé príklady (zoradené)}
		\begin{tabular}{|l|r|r|}
		\hline
		\textbf{Príklad} & \textbf{Absolútna} & \textbf{Relatívna (\%)} \\ \hline
		`)
		for _, q := range questions {
			builder.WriteString(fmt.Sprintf("%d & %d & %.2f \\\\\n\\hline\n", q.number, q.absolute, q.relative))
		}
		builder.WriteString(`\end{tabular}`)
	}

	builder.WriteString(`\end{document}`)

	return []byte(builder.String()), nil
}

// buildPlotData processes the score data into labels and coordinates for LaTeX plotting.
func buildPlotData(scores []int) (labels string, coordinates string) {
	distribution := make(map[int]int)
	for _, s := range scores {
		lb := (s / 10) * 10
		distribution[lb]++
	}

	var lbs []int
	for lb := range distribution {
		lbs = append(lbs, lb)
	}
	sort.Ints(lbs)

	var labelParts, coordParts []string
	for _, lb := range lbs {
		labelParts = append(labelParts, fmt.Sprintf("%d-%d", lb, lb+9))
		coordParts = append(coordParts, fmt.Sprintf("(%d,%d)", lb, distribution[lb]))
	}

	return strings.Join(labelParts, ","), strings.Join(coordParts, "\n")
}

// buildPerQuestionPlotData processes the success rate data for each question.
func buildPerQuestionPlotData(successRates []float64) (labels string, coordinates string) {
	var labelParts, coordParts []string
	for i, rate := range successRates {
		labelParts = append(labelParts, fmt.Sprintf("%d", i+1))              // Popisky pre os y (príklady)
		coordParts = append(coordParts, fmt.Sprintf("(%.1f,%d)", rate, i+1)) // Opravené súradnice: (úspešnosť, príklad)
	}
	return strings.Join(labelParts, ","), strings.Join(coordParts, "\n")
}

// calculateSuccessPerQuestion calculates the success rate per question.
// It compares each student answer to the correct answers and computes the percentage of correct answers per question.
func calculateSuccessPerQuestion(students []models.Student, correctAnswers string, totalQuestions int) []float64 {
	successRates := make([]float64, totalQuestions)

	for q := 0; q < totalQuestions; q++ {
		correctCount := 0
		for _, s := range students {
			if len(s.Answers) > q && s.Answers[q] == correctAnswers[q] {
				correctCount++
			}
		}
		successRates[q] = float64(correctCount) / float64(len(students)) * 100
	}

	return successRates
}

// calculatePerQuestionSuccess calculates both the absolute and relative success per question.
// It returns two slices: one with the absolute number of correct answers per question,
// and one with the relative percentage of success per question.
func calculatePerQuestionSuccess(students []models.Student, correctAnswers string, totalQuestions int) ([]int, []float64) {
	absolute := make([]int, totalQuestions)
	relative := make([]float64, totalQuestions)
	totalStudents := len(students)

	for q := 0; q < totalQuestions; q++ {
		correctCount := 0
		for _, s := range students {
			if len(s.Answers) > q && s.Answers[q] == correctAnswers[q] {
				correctCount++
			}
		}
		absolute[q] = correctCount
		relative[q] = float64(correctCount) / float64(totalStudents) * 100
	}

	return absolute, relative
}
