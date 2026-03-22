package seed

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"
	"log/slog"
	"math/rand"
)

func StudentGenerator(questionsCount int) *models.Student {
	firstnames := []string{"Ján", "František", "Jozef", "Martin", "Monika", "Nikola", "Ema", "Vanesa", "Timotej", "Matúš", "Roman Alexander", "Radoslav", "Ondrej"}
	surnames := []string{"Novák", "Kováč", "Horvát", "Štúr", "Nagy", "Varga", "Kolesár", "Mrkvička", "Kokavec", "Matejovec", "Šeliga"}
	rooms := []string{"AB300", "BC300", "CD300", "DE300", "AB150"}

	student := &models.Student{
		Name:               firstnames[rand.Intn(len(firstnames))],
		Surname:            surnames[rand.Intn(len(surnames))],
		BirthDate:          RandomDate(),
		RegistrationNumber: generateRegistrationNumber(),
		Room:               rooms[rand.Intn(len(rooms))],
		Score:              0,
		Answers:            GenerateAnswers(questionsCount),
	}

	return student
}

func StudentListGenerator(questionsCount int, studentsCount int) *[]models.Student {
	logger := logging.GetLogger()

	logger.Info("Generovanie študentov...", slog.Int("count", studentsCount))
	students := []models.Student{}
	for i := 0; i < studentsCount; i++ {
		students = append(students, *StudentGenerator(questionsCount))
	}
	return &students
}

func generateRegistrationNumber() int {
	n := rand.Intn(999999)
	for n < 100000 {
		n = rand.Intn(999999)
	}
	return n
}
