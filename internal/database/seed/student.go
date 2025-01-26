package seed

import (
	"ScanEvalApp/internal/database/models"
	"math/rand"
	"strconv"
	"time"
)

func StudentGenerator(questionsCount int) *models.Student {
	firstnames := []string{"Ján", "František", "Jozef", "Martin", "Monika", "Nikola", "Ema", "Vanesa", "Timotej", "Matúš", "Roman Alexander", "Radoslav", "Ondrej"}
	surnames := []string{"Novák", "Kováč", "Horvát", "Štúr", "Nagy", "Varga", "Kolesár", "Mrkvička", "Kokavec", "Matejovec", "Šeliga"}
	rooms := []string{"AB300", "BC300", "CD300", "DE300", "AB150"}

	student := &models.Student{
		Name:               firstnames[rand.Intn(len(firstnames))],
		Surname:            surnames[rand.Intn(len(surnames))],
		BirthDate:          randomDate(),
		RegistrationNumber: generateRegistrationNumber(),
		Room:               rooms[rand.Intn(len(rooms))],
		Score:              0,
		Answers:            generateAnswers(questionsCount),
	}

	return student
}

func StudentListGenerator(questionsCount int) *[]models.Student {
	students := []models.Student{}
	for i := 0; i < 50; i++ {
		students = append(students, *StudentGenerator(questionsCount))
	}
	return &students
}

func generateRegistrationNumber() string {
	s := strconv.FormatInt(int64(rand.Intn(99999999)), 10)
	for len(s) < 8 {
		s = "0" + s
	}
	return s
}

func randomDate() time.Time {
	minUnix := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	maxUnix := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	randomUnix := rand.Int63n(maxUnix-minUnix) + minUnix
	return time.Unix(randomUnix, 0)
}
