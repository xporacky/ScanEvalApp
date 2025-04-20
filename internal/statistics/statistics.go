package statistics

import (
	"ScanEvalApp/internal/database/models"
)

// generateStatistics generuje štatistiky podľa vybraných možností
func GenerateStatistics(selectedStats []string, exam *models.Exam) {
	println("štatistiky pre test: ", exam.Title)
	// Tento bod by mal byť nahradený logikou generovania štatistík podľa vybraných možností.
	for _, stat := range selectedStats {
		// Spracovanie vybranej štatistiky, napríklad:
		// - Maximum bodov
		// - Minimum bodov
		// - Priemer
		// atď.
		// Tu môžeš pridať ďalšie spracovanie alebo volanie funkcií na generovanie konkrétnych štatistík.
		println("Generovanie štatistiky:", stat)
	}
}
