package latex

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/config"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenerateStatistics generates statistics based on selected options
// and compiles a LaTeX report for the exam data.
// It returns the file path of the generated PDF or an error.
func GenerateStatistics(selectedStats []string, exam *models.Exam) (string, error) {
	errorLogger := logging.GetErrorLogger()

	students := filterParticipatingStudents(exam.Students)

	scores := getScores(students)
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
	statsData["includeRoomStats"] = false
	statsData["includeGroupStats"] = false
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
			successPerQuestion := calculateSuccessPerQuestion(students, exam)
			statsData["includePerQuestionDistribution"] = true
			statsData["successPerQuestion"] = successPerQuestion

		case "Úspešnosť absolútna aj relatívna":
			absolute, relative := calculateOverallSuccess(students, exam.QuestionCount)
			statsData["includeOverallSuccess"] = true
			statsData["sumPoints"] = exam.QuestionCount * len(students)
			statsData["absoluteSuccess"] = absolute
			statsData["relativeSuccess"] = relative * 100

		case "Úspešnosť absolútna aj relatívna pre jednotlivé príklady":
			absolutePerQuestion, relativePerQuestion := calculatePerQuestionSuccess(students, exam)
			statsData["includePerQuestionSuccess"] = true
			statsData["absolutePerQuestion"] = absolutePerQuestion
			statsData["relativePerQuestion"] = relativePerQuestion

		case "Štatistika podľa miestnosti":
			groups := groupStudentsByRoom(students, exam)
			statsData["includeRoomStats"] = true
			statsData["roomGroups"] = groups

		case "Štatistika podľa skupín":
			subgroups := groupStudentsBySubgroup(students)
			statsData["includeGroupStats"] = true
			statsData["subgroups"] = subgroups

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

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
		return "", err
	}

	outputPath := filepath.Join(absDirPath, fmt.Sprintf("stats_%s.pdf", common.SanitizeFilename(exam.Title)))
	if err := os.WriteFile(outputPath, pdfBytes, common.FILE_PERMISSION); err != nil {
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

// filterParticipatingStudents keeps only students whose answer sheet was processed.
// Pages is set during scan processing even when the student scores 0 points.
func filterParticipatingStudents(students []models.Student) []models.Student {
	participating := make([]models.Student, 0, len(students))
	for _, s := range students {
		if strings.TrimSpace(s.Pages) != "" {
			participating = append(participating, s)
		}
	}
	return participating
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
	if totalPossible == 0 {
		return totalCorrect, 0
	}
	relative := float64(totalCorrect) / float64(totalPossible)
	return totalCorrect, relative
}

// latexEscape escapes special characters for LaTeX compatibility.
func latexEscape(str string) string {
	replacer := strings.NewReplacer(
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
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
	return strings.TrimSpace(replacer.Replace(str))
}

// RoomGroup represents a group of students identified by their date, time, and room.
type RoomGroup struct {
	Key      string
	Students []models.Student
}

// groupStudentsBySubgroup groups students by their Subgroup field and returns sorted groups.
func groupStudentsBySubgroup(students []models.Student) []RoomGroup {
	groupMap := make(map[string][]models.Student)
	for _, s := range students {
		key := strings.TrimSpace(s.Subgroup)
		if key == "" {
			key = "(bez skupiny)"
		}
		groupMap[key] = append(groupMap[key], s)
	}
	keys := make([]string, 0, len(groupMap))
	for k := range groupMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	groups := make([]RoomGroup, 0, len(keys))
	for _, k := range keys {
		groups = append(groups, RoomGroup{Key: k, Students: groupMap[k]})
	}
	return groups
}

// groupStudentsByRoom groups students by their (date, time, room) triple and returns sorted groups.
func groupStudentsByRoom(students []models.Student, exam *models.Exam) []RoomGroup {
	groupMap := make(map[string][]models.Student)
	for _, s := range students {
		date := exam.Date.Format("02.01.2006")
		if !s.ExamDate.IsZero() {
			date = s.ExamDate.Format("02.01.2006")
		}
		timeStr := exam.Date.Format("15:04")
		if s.ExamTime != "" {
			timeStr = s.ExamTime
		}
		key := date + ", " + timeStr + ", " + s.Room
		groupMap[key] = append(groupMap[key], s)
	}
	keys := make([]string, 0, len(groupMap))
	for k := range groupMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	groups := make([]RoomGroup, 0, len(keys))
	for _, k := range keys {
		groups = append(groups, RoomGroup{Key: k, Students: groupMap[k]})
	}
	return groups
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
	\pagestyle{empty}
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
		labels, coords := buildPlotData(statsData["scores"].([]int), exam.QuestionCount)

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

	// Room statistics pages
	if statsData["includeRoomStats"].(bool) {
		groups := statsData["roomGroups"].([]RoomGroup)

		for _, group := range groups {
			if len(group.Students) == 0 {
				continue
			}

			groupScores := getScores(group.Students)
			groupMax := calculateMax(groupScores)
			groupMin := calculateMin(groupScores)
			groupAvg := calculateAverage(groupScores)
			groupMedian := calculateMedian(groupScores)
			groupAbsolute, groupRelative := calculateOverallSuccess(group.Students, exam.QuestionCount)

			builder.WriteString("\n\\clearpage\n\\vspace*{-1cm}\n")
			builder.WriteString("\\begin{center}\n{\\large\\textbf{Štatistiky miestnosti: " + latexEscape(group.Key) + "}}\n\\end{center}\n")
			builder.WriteString("\\vspace{0.4cm}\n\n")

			builder.WriteString("\\begin{center}\n\\textbf{Základné štatistiky}\\\\[0.2em]\n")
			builder.WriteString("\\begin{tabular}{|l|r|}\n\\hline\n")
			builder.WriteString(fmt.Sprintf("Počet študentov & %d \\\\\n\\hline\n", len(group.Students)))
			builder.WriteString(fmt.Sprintf("Maximum bodov & %d \\\\\n\\hline\n", groupMax))
			builder.WriteString(fmt.Sprintf("Minimum bodov & %d \\\\\n\\hline\n", groupMin))
			builder.WriteString(fmt.Sprintf("Priemer & %.2f \\\\\n\\hline\n", groupAvg))
			builder.WriteString(fmt.Sprintf("Medián & %.2f \\\\\n\\hline\n", groupMedian))
			builder.WriteString(fmt.Sprintf("Absolútna úspešnosť & %d \\\\\n\\hline\n", groupAbsolute))
			builder.WriteString(fmt.Sprintf("Relatívna úspešnosť & %.2f\\%% \\\\\n\\hline\n", groupRelative*100))
			builder.WriteString("\\end{tabular}\n\\end{center}\n\n")

			groupLabels, groupCoords := buildPlotData(groupScores, exam.QuestionCount)
			builder.WriteString("\\vspace{0.4cm}\n\n")
			builder.WriteString("\\noindent\\textbf{Rozdelenie bodov}\n")
			builder.WriteString(`
	\begin{tikzpicture}
	\begin{axis}[
		ybar,
		xlabel={Rozsahy bodov},
		ylabel={Počet študentov},
		width=\textwidth,
		height=5cm,
		bar width=0.6cm,
		xtick=data,
		xticklabels={` + groupLabels + `},
		nodes near coords,
	]
	\addplot coordinates {
	` + groupCoords + `
	};
	\end{axis}
	\end{tikzpicture}
	`)

			if exam.QuestionCount > 0 {
				absPerQ, relPerQ := calculatePerQuestionSuccess(group.Students, exam)
				builder.WriteString("\n\\vspace{0.4cm}\n\n")
				builder.WriteString("\\begin{center}\n\\textbf{Úspešnosť za jednotlivé príklady}\\\\[0.2em]\n")
				builder.WriteString("{\\footnotesize\n")
				builder.WriteString("\\begin{tabular}{|l|r|r|}\n\\hline\n")
				builder.WriteString("\\textbf{Príklad} & \\textbf{Absolútna} & \\textbf{Relatívna (\\%)} \\\\\n\\hline\n")
				for q := 0; q < len(absPerQ); q++ {
					builder.WriteString(fmt.Sprintf("%d & %d & %.2f \\\\\n\\hline\n", q+1, absPerQ[q], relPerQ[q]))
				}
				builder.WriteString("\\end{tabular}\n}\n\\end{center}\n")
			}
		}
	}

	// Group (subgroup) statistics pages
	if statsData["includeGroupStats"].(bool) {
		subgroups := statsData["subgroups"].([]RoomGroup)

		for _, group := range subgroups {
			if len(group.Students) == 0 {
				continue
			}

			groupScores := getScores(group.Students)
			groupMax := calculateMax(groupScores)
			groupMin := calculateMin(groupScores)
			groupAvg := calculateAverage(groupScores)
			groupMedian := calculateMedian(groupScores)
			groupAbsolute, groupRelative := calculateOverallSuccess(group.Students, exam.QuestionCount)

			builder.WriteString("\n\\clearpage\n\\vspace*{-1cm}\n")
			builder.WriteString("\\begin{center}\n{\\large\\textbf{Štatistiky skupiny: " + latexEscape(group.Key) + "}}\n\\end{center}\n")
			builder.WriteString("\\vspace{0.4cm}\n\n")

			builder.WriteString("\\begin{center}\n\\textbf{Základné štatistiky}\\\\[0.2em]\n")
			builder.WriteString("\\begin{tabular}{|l|r|}\n\\hline\n")
			builder.WriteString(fmt.Sprintf("Počet študentov & %d \\\\\n\\hline\n", len(group.Students)))
			builder.WriteString(fmt.Sprintf("Maximum bodov & %d \\\\\n\\hline\n", groupMax))
			builder.WriteString(fmt.Sprintf("Minimum bodov & %d \\\\\n\\hline\n", groupMin))
			builder.WriteString(fmt.Sprintf("Priemer & %.2f \\\\\n\\hline\n", groupAvg))
			builder.WriteString(fmt.Sprintf("Medián & %.2f \\\\\n\\hline\n", groupMedian))
			builder.WriteString(fmt.Sprintf("Absolútna úspešnosť & %d \\\\\n\\hline\n", groupAbsolute))
			builder.WriteString(fmt.Sprintf("Relatívna úspešnosť & %.2f\\%% \\\\\n\\hline\n", groupRelative*100))
			builder.WriteString("\\end{tabular}\n\\end{center}\n\n")

			groupLabels, groupCoords := buildPlotData(groupScores, exam.QuestionCount)
			builder.WriteString("\\vspace{0.4cm}\n\n")
			builder.WriteString("\\noindent\\textbf{Rozdelenie bodov}\n")
			builder.WriteString(`
	\begin{tikzpicture}
	\begin{axis}[
		ybar,
		xlabel={Rozsahy bodov},
		ylabel={Počet študentov},
		width=\textwidth,
		height=5cm,
		bar width=0.6cm,
		xtick=data,
		xticklabels={` + groupLabels + `},
		nodes near coords,
	]
	\addplot coordinates {
	` + groupCoords + `
	};
	\end{axis}
	\end{tikzpicture}
	`)

			if exam.QuestionCount > 0 {
				absPerQ, relPerQ := calculatePerQuestionSuccess(group.Students, exam)
				builder.WriteString("\n\\vspace{0.4cm}\n\n")
				builder.WriteString("\\begin{center}\n\\textbf{Úspešnosť za jednotlivé príklady}\\\\[0.2em]\n")
				builder.WriteString("{\\footnotesize\n")
				builder.WriteString("\\begin{tabular}{|l|r|r|}\n\\hline\n")
				builder.WriteString("\\textbf{Príklad} & \\textbf{Absolútna} & \\textbf{Relatívna (\\%)} \\\\\n\\hline\n")
				for q := 0; q < len(absPerQ); q++ {
					builder.WriteString(fmt.Sprintf("%d & %d & %.2f \\\\\n\\hline\n", q+1, absPerQ[q], relPerQ[q]))
				}
				builder.WriteString("\\end{tabular}\n}\n\\end{center}\n")
			}
		}
	}

	builder.WriteString(`\end{document}`)

	return []byte(builder.String()), nil
}

// buildPlotData produces exactly 5 bins for a score distribution histogram.
// Bin 0: students with exactly 0 points.
// Bins 1-4: the range [1..maxScore] split into 4 equal parts (ceiling division).
func buildPlotData(scores []int, maxScore int) (labels string, coordinates string) {
	bins := make([]int, 5)
	binSize := 1
	if maxScore > 0 {
		binSize = (maxScore + 3) / 4 // ceil(maxScore / 4)
	}

	for _, s := range scores {
		if s <= 0 {
			bins[0]++
		} else {
			b := (s-1)/binSize + 1
			if b > 4 {
				b = 4
			}
			bins[b]++
		}
	}

	var labelParts, coordParts []string
	labelParts = append(labelParts, "0")
	coordParts = append(coordParts, fmt.Sprintf("(0,%d)", bins[0]))
	for i := 1; i <= 4; i++ {
		lo := (i-1)*binSize + 1
		hi := i * binSize
		if maxScore > 0 && hi > maxScore {
			hi = maxScore
		}
		if lo >= hi {
			labelParts = append(labelParts, fmt.Sprintf("%d", lo))
		} else {
			labelParts = append(labelParts, fmt.Sprintf("%d-%d", lo, hi))
		}
		coordParts = append(coordParts, fmt.Sprintf("(%d,%d)", i, bins[i]))
	}

	return strings.Join(labelParts, ","), strings.Join(coordParts, "\n")
}

// buildPerQuestionPlotData processes the success rate data for each question.
func buildPerQuestionPlotData(successRates []float64) (labels string, coordinates string) {
	var labelParts, coordParts []string
	for i, rate := range successRates {
		labelParts = append(labelParts, fmt.Sprintf("%d", i+1))
		coordParts = append(coordParts, fmt.Sprintf("(%.1f,%d)", rate, i+1))
	}
	return strings.Join(labelParts, ","), strings.Join(coordParts, "\n")
}

// calculateSuccessPerQuestion calculates the success rate per question.
// For multi-day exams each student's correct answers are determined by their subgroup.
func calculateSuccessPerQuestion(students []models.Student, exam *models.Exam) []float64 {
	totalQuestions := exam.QuestionCount
	successRates := make([]float64, totalQuestions)
	if len(students) == 0 {
		return successRates
	}

	var subgroupAnswers map[string]string
	if exam.IsMultiDay {
		subgroupAnswers = parseSubgroupAnswers(exam.Questions)
	}

	for q := 0; q < totalQuestions; q++ {
		correctCount := 0
		for _, s := range students {
			correct := getStudentCorrectAnswers(s, exam, subgroupAnswers)
			if len(s.Answers) > q && len(correct) > q && s.Answers[q] == correct[q] {
				correctCount++
			}
		}
		successRates[q] = float64(correctCount) / float64(len(students)) * 100
	}

	return successRates
}

// calculatePerQuestionSuccess calculates both the absolute and relative success per question.
// For multi-day exams each student's correct answers are determined by their subgroup.
func calculatePerQuestionSuccess(students []models.Student, exam *models.Exam) ([]int, []float64) {
	totalQuestions := exam.QuestionCount
	absolute := make([]int, totalQuestions)
	relative := make([]float64, totalQuestions)
	totalStudents := len(students)
	if totalStudents == 0 {
		return absolute, relative
	}

	var subgroupAnswers map[string]string
	if exam.IsMultiDay {
		subgroupAnswers = parseSubgroupAnswers(exam.Questions)
	}

	for q := 0; q < totalQuestions; q++ {
		correctCount := 0
		for _, s := range students {
			correct := getStudentCorrectAnswers(s, exam, subgroupAnswers)
			if len(s.Answers) > q && len(correct) > q && s.Answers[q] == correct[q] {
				correctCount++
			}
		}
		absolute[q] = correctCount
		relative[q] = float64(correctCount) / float64(totalStudents) * 100
	}

	return absolute, relative
}

// parseSubgroupAnswers parses the JSON-encoded subgroup answers from an exam's Questions field.
// Returns nil if the content is not valid JSON (e.g. plain answer string for non-multi-day exams).
func parseSubgroupAnswers(questionsJSON string) map[string]string {
	var result map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(questionsJSON)), &result); err != nil {
		return nil
	}
	return result
}

// getStudentCorrectAnswers returns the lowercased correct answer string for a given student.
// For multi-day exams it looks up the student's subgroup in the parsed answers map.
func getStudentCorrectAnswers(s models.Student, exam *models.Exam, subgroupAnswers map[string]string) string {
	if subgroupAnswers != nil {
		if answers, ok := subgroupAnswers[s.Subgroup]; ok {
			return strings.ToLower(answers)
		}
		return ""
	}
	return strings.ToLower(strings.TrimSpace(exam.Questions))
}
