package tests

import (
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/logging"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var expectedResults = map[int]string{
	2:  "acdcebcdbcdcbecbbcbdacbcdedecdcabcbcdedc",
	3:  "acbbexcbdbaabcccbedeabbccceadbdcebcdbced",
	32: "bcadabdcabecbcaedabdbcdabdeabdbcdaeadcbd",
	27: "acxebbbcbbcdeaabccaebecbdabcdcaabbcbcdcb",
	22: "abaceabcxadbeaeabcdecccccbeeaabcadaedabd",
	35: "dcbaedcbaedcbaedcbaebacecabccecabdcbdaec",
	13: "cdbdcbbcdddccbcabdbadbadedbcedbcbcdxbcea",
	24: "abdddeebacbaebdabcedabeddcbaedabecdabccc",
	47: "cedbabcxddbabdeabcdaabcxeccbxebbbdccccba",
	14: "abcededbacabcdeddddddbceaabcedaedbccbdae",
	49: "acbdcbcdccabdcbeccbeecdbcadcbdedddccceba",
	18: "abceedcebaabedcabcdebacddedcbaabcdeaaaaa",
	15: "cdbeacbaedacedbabcdeabecdabcedbbbbbcdeab",
	44: "dcdbeabcccddbcbadcbaaccbdedacdaedcdcbbxa",
	45: "aedbcacebdabdecbcaeaaeaeabdbdbcacacdcdcd",
	43: "abccbcdcdeebcecdbebaabcdxdccaecbcebbddee",
	41: "abcdaedecabcdeabacdeeeeeeabcdebdacebdeac",
	8:  "addcaabecxbcddcdbacaabbbcbbccdabebeedcec",
	29: "abcdebbbbbcccccdddddabcdedacebabdecbacde",
	21: "abcebcdccaabccebbcdecacddebaxabcccdedcba",
	12: "edcdddcbbaedccbbedddabbcdcccbbbcddcecdba",
	38: "dabcedecababceeabcdexxxxxxxxxxxxxxxxxxxx",
	7:  "abdceabceabcdeabacdeabceeabcdeabcdeaaaaa",
	4:  "edcbxcebdabcdbxadcdaabdcdcccccdacbebcbde",
	20: "abcedabcddedcabacdeabaedcabcdceabdeacbed",
	34: "edcddcbdcbeddbcacbbebccaedecbcbcbdcabcdd",
	46: "abecedccbacbbddbddcaeddcbeedcxacbxeabced",
	26: "aabcecdbaaeeeecddabeabccddceabacedabcdea",
	39: "ebcaebcbaecbbdccabdcabxbdbdcbxcdbcaabxce",
	9:  "cbdbeabecxabdeabcdcexxxxxxxxxxxxxxxxxxxx",
	16: "abcdeabcdeabcdedceabaabcceecdbcabecbceab",
	6:  "edcbabcbcecbdaeecbababcceabcbabbbdeccbae",
	42: "abbbbbbbacdedbabcdcexxxxxxxxxxxxxxxxxxxx",
	5:  "acdcecbadbdcbccabbeeaaccecabceabaadaddce",
	36: "abcaedcabdcbdaedbacdeeeeedcxabcdeabcaade",
	25: "bcaedabcdeabcdeabbccdddddbbbbbcdaceabced",
	23: "abcdceedcdabdeabcdcexabcecbcabacdeabecbc",
	17: "acbdeabcedabcedabdceabdcabcdaaxcaedbadca",
	40: "abababcdbcdbcdbcdeababceacbdaeabdecabdca",
	19: "abcdaabcaabcabeddeddababacdcdcxdededcbac",
	31: "bcaedbacdedebacbaacecdbeabcedabcdeabdcae",
	1:  "abcdeabcdeedcbaabcdeabcdeedcbaabcdeedcba",
	37: "ebadecbdabcdbaebcdccabcbdabcdabcdeabcdea",
	28: "abdccedcbcbbbbbdededaedbcebacedbceabcdea",
	10: "aadcebecdbabeabcxbecaadecbabedcdbadcdcba",
	11: "cdaebdcacxdbceeabbccbbdddaaeedcbabdedcbb",
	50: "bcaxedcbceaaaaabcdddabcddcbbcabdcdeeaabb",
	30: "dcdcedabcdbacceaacaaedcbabcdeeddccbbaaee",
	48: "eddcdxaabcbabacadaedbcaedcdebaabacaeddcb",
	33: "baacecbdxbcbdaaeecdabbbdacedabcbaecabcce",
}

var expectedCorrectedResults = map[int]map[int]string{
	13: {
		1:  "c",
		15: "c",
		17: "c",
		33: "b",
	},
	47: {
		9:  "d",
		10: "d",
		13: "b",
		25: "e",
		30: "e",
		37: "c",
	},
	14: {
		6:  "e",
		21: "d",
	},
	49: {
		8:  "d",
		9:  "c",
		10: "c",
		17: "c",
		34: "d",
	},
	44: {
		19: "b",
		22: "c",
		24: "b",
		25: "d",
		33: "d",
	},
	43: {
		14: "e",
		18: "e",
		29: "a",
		36: "b",
	},
	8: {
		2:  "d",
		23: "b",
		39: "e",
	},
	21: {
		13: "c",
		23: "b",
		39: "e",
	},
	12: {
		17: "e",
	},
	7: {
		4: "c",
	},
	4: {
		15: "e",
	},
	34: {
		6:  "c",
		27: "e",
	},
	46: {
		29: "c",
	},
	39: {
		34: "c",
		35: "a",
	},
	16: {
		29: "d",
	},
	6: {
		24: "c",
		26: "a",
	},
	42: {
		9: "a",
	},
	5: {
		4:  "c",
		15: "c",
	},
	23: {
		9: "c",
	},
	37: {
		25: "d",
	},
	10: {
		13: "e",
	},
	48: {
		11: "b",
	},
	33: {
		24: "d",
	},
}

func setupTestDB() (*gorm.DB, error) {
	testDBPath := "../internal/database/scan-eval-test-db.db"
	db, err := gorm.Open(sqlite.Open(testDBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestAnswerRecognition(t *testing.T) {
	errorLogger := logging.GetErrorLogger()
	db, err := setupTestDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		t.Fatalf("Nepodarilo sa pripojiť k databáze: %v", err)
	}

	totalQuestions := 0
	totalCorrect := 0
	totalMissing := 0
	totalUnrecognized := 0 // New counter for unrecognized answers

	for studentID, expectedAnswers := range expectedResults {
		student, err := repository.GetStudentById(db, uint(studentID), 1) // testID = 1
		if err != nil {
			t.Errorf("Študent %d nebol nájdený: %v", studentID, err)
			totalMissing += 40
			continue
		}

		recognizedAnswers := student.Answers
		if len(recognizedAnswers) == 0 {
			t.Errorf("Študent %d: chýbajúce odpovede", studentID)
			totalMissing += 40
			continue
		}

		correctCount := 0
		missingCount := 0
		unrecognized := 0
		totalQuestions += 40

		for i := 0; i < 40; i++ {
			if i >= len(recognizedAnswers) {
				t.Errorf("Študent %d, otázka %d: chýbajúca odpoveď", studentID, i+1)
				missingCount++
				continue
			}

			if recognizedAnswers[i] == expectedAnswers[i] {
				correctCount++
			} else if recognizedAnswers[i] == '0' { // Check for unrecognized answer
				totalUnrecognized++
				unrecognized++
			}
			//t.Errorf("Študent %d, otázka %d: OCR nezachytilo odpoveď", studentID, i+1)
			// } else {
			// 	t.Errorf("Študent %d, otázka %d: očakávané %c, rozpoznané %c",
			// 		studentID, i+1, expectedAnswers[i], recognizedAnswers[i])
			// }
		}

		totalCorrect += correctCount
		totalMissing += missingCount
		fmt.Printf("Študent %d: správne %d/40, chýbajúce %d, nezachytené %d\n", studentID, correctCount, missingCount, unrecognized)

	}

	successRate := float64(totalCorrect) / float64(totalQuestions) * 100
	fmt.Printf("Celková úspešnosť OCR: %.2f%% (%d/%d správnych odpovedí, %d chýbajúcich, %d nezachytených)\n",
		successRate, totalCorrect, totalQuestions, totalMissing, totalUnrecognized)
}

func TestStudentAnswersExistence(t *testing.T) {
	errorLogger := logging.GetErrorLogger()
	db, err := setupTestDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		t.Fatalf("Nepodarilo sa pripojiť k databáze: %v", err)
	}
	totalStudents := len(expectedResults)
	recognizedCount := 0
	for studentID := range expectedResults {
		student, err := repository.GetStudentById(db, uint(studentID), 1) // testID = 1
		if err != nil {
			t.Errorf("Študent %d nebol nájdený: %v", studentID, err)
			continue
		}
		recognizedCount++

		answers := student.Answers

		count := strings.Count(answers, "0")
		recognizedAnswers := len(answers) - count
		fmt.Printf("Študent %d: rozpoznané odpovede %d/%d\n", studentID, recognizedAnswers, len(answers))

	}
	fmt.Printf("Celkový počet študentov: %d, rozpoznaných študentov: %d\n", totalStudents, recognizedCount)

}

func TestMissingPages(t *testing.T) {
	errorLogger := logging.GetErrorLogger()
	db, err := setupTestDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		t.Fatalf("Nepodarilo sa pripojiť k databáze: %v", err)
	}

	missingPages := 0

	for studentID := range expectedResults {
		student, err := repository.GetStudentById(db, uint(studentID), 1) // testID = 1
		if err != nil {
			t.Errorf("Študent %d nebol nájdený: %v", studentID, err)
			missingPages++
			missingPages++
			continue
		}

		recognizedAnswers := student.Answers

		// Počítaj počet znakov '0' v odpovediach
		zeroCount := strings.Count(recognizedAnswers, "0")

		// Ak je počet znakov '0' 20 alebo viac, znamená to, že chýba aspoň jedna strana
		if zeroCount == 40 {
			missingPages += 2 // Zaznamenaj jednu chýbajúcu stranu
		}
		if zeroCount == 20 {
			missingPages++
		}
	}

	fmt.Printf("Celkový počet chýbajúcich strán: %d\n", missingPages)
}

func TestCorrectedAnswerRecognition(t *testing.T) {
	errorLogger := logging.GetErrorLogger()
	db, err := setupTestDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		t.Fatalf("Nepodarilo sa pripojiť k databáze: %v", err)
	}

	totalCorrect := 0
	totalQuestions := 0
	totalUnrecognized := 0

	for studentID, expectedAnswers := range expectedCorrectedResults {
		student, err := repository.GetStudentById(db, uint(studentID), 1) // testID = 1
		if err != nil {
			t.Errorf("Študent %d nebol nájdený: %v", studentID, err)
			continue
		}

		recognizedAnswers := student.Answers
		if len(recognizedAnswers) == 0 {
			t.Errorf("Študent %d: chýbajúce odpovede", studentID)
			totalUnrecognized += len(expectedAnswers)
			continue
		}

		for questionID, expectedAnswer := range expectedAnswers {
			totalQuestions++

			recognizedAnswer := string(recognizedAnswers[questionID-1]) // Convert byte to string

			if recognizedAnswer == expectedAnswer {
				totalCorrect++
			} else if recognizedAnswer == "0" { // Check for unrecognized answer
				totalUnrecognized++
				t.Errorf("Študent %d, otázka %d: OCR nezachytilo odpoveď", studentID, questionID)
			} else {
				t.Errorf("Študent %d, otázka %d: očakávané %s, rozpoznané %s", studentID, questionID, expectedAnswer, recognizedAnswer)
			}
		}
	}

	successRate := float64(totalCorrect) / float64(totalQuestions) * 100
	fmt.Printf("Celková úspešnosť rozpoznania opravených odpovedí: %.2f%% (%d/%d správnych odpovedí, %d nezachytených)\n",
		successRate, totalCorrect, totalQuestions, totalUnrecognized)
}
