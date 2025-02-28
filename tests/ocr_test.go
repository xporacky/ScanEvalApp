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

var expectedAnswer_1_page = map[int]string{
	392411: "bccdeeddccdddcxcxceddcxebxcceddbeababddc",
	913602: "aaaaabbbbbcccccdddddbcbcdbcdbcbababxxxxx",
	985335: "abdcxbcdcdbxxaedxdcbabccbcbcdecbabcccbbc",
	257457: "abdcaxdeccbcxaebbxecabcccccccbdcbadxaccd",
	133168: "abcdeacdedxbdcbcdcdcbcdedcbacedcbabdedcb",
	526606: "abcbacbcdcbcdbcbdxcxbbbbbxbbbxbbbcdxcxbc",
	411098: "bdbcdbaaaabbbbbcccccdacdbcxdxdcxdcbdbxce",
	631788: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	433835: "bedcdcdeccbeadcxxxxxaaabeabbbcbcdecaaaaa",
	300532: "abcccbbcccbbbbbdddddeeeeeeeeeedddddaaaaa",
	801650: "abbbbddddddcxabadcdbaaaaabbbbbcccccbbbbb",
	783424: "abccccccccbbbbbacbcdabbbbccccceeeeeaaaaa",
	990337: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	753491: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	648037: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	188319: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	988443: "xxxxxxxxxxxxxxcxxxxxabcdexxxxxxxxxxxxxxx",
	624245: "abcdeeeeeeeeeeeeeeeexdxxxexdeeeeeeeadxed",
	774776: "xdaaaaaaaaaaabbbcdeeabcddeeeeeeeeeeeeeee",
	236273: "aaaaaaaaaaaaaxaaaaaaaaaaaaaaaaaaaaaaaaaa",
	537282: "cccccccccccccccccccccccccccccccccccccccc",
	639863: "cccccccccccccccccccccccccccccccccccccccc",
	227633: "cccccccccccccccccccccccccccccccccccccccc",
	872413: "cccccccccccccccccccccccccccccccccccccccc",
	212971: "abcxcdxcxcdeddcecdedaxcdebbddedddedxcccd",
	507113: "xcccxxbdddddddccccxbcccccccccccccccccccc",
	152342: "dddddddddddddddddddddddddddddddddddddddd",
	870991: "abbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	783456: "ccccccccccccccccccccdddcccbbbbbbbaxxxdec",
	590787: "bbbbbbbbbbbbbbbbbbbbcccccccccccccccccccc",
	705932: "abababcbcbcdcdcdededabababcbcbcdcdcdeded",
	363580: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	168030: "bdcaabccccdcccccccccedcbaabcdedcbaaabxde",
	646042: "bdecacdecccbdccxxbccaaaaaaaaaaaaaaaaaaaa",
	375942: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	658013: "bcdecbcdedbcdddbbbbbbbbbbcbbbbbbbbbbbbbb",
	798399: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	984996: "abcdebcdeeabcdeabcdeabcdeabcdeabcdeabcde",
	484470: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxaaaxxxxxxxx",
	731579: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	235850: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	404065: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	760018: "ccccccccccccccccccccxxxxxbccdddddddddddd",
	422417: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	636975: "bbbbbbbbbbbbbbbbbbbbxxxxxxxxxxxxxxxxxxxx",
	567742: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	123800: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	385676: "bbbbbbbbbbbbbbbbbbbbbbbbbcccccbbbbbaaaaa",
	761336: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	110198: "dabcdcccccdeexdabcdeabcdeabcdeedcbbabxed",	
}

var expectedAnswer_both_page = map[int]string{
	507113: "babababababababababababababababababababx",	
	227633: "xxxxccccccccccccccccxxxxxxxxxxxxxxxxxccc",	
	152342: "cccccccccccccccccccccccccccccccccccccccc",	
	870991: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",	
	590787: "dddddddddddddddddddddddddddddddddddddddd",	
	646042: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",	
	984996: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",	
	404065: "dddddddddddddddddddddddddddddddddddddddd",	
	567742: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	385676: "abbcxabcdexxxxxabcdeabcdeabcdeabxxeabcde",	
	774776: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	375942: "aaaxxxxbbbbbbbbbbbbbxxxxxxxbxxxxxxbxxbxx",	
	433835: "dddddddddddddddddddddddddddddddddddddddd",	
	411098: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	526606: "bbbbbbbbbbbbbbbbbbbbccccccccccccccbbbbbb",	
	257457: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	783456: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",	
	801650: "ebccbcexcdxecbdbbbcdaaaaaadcbccdaabccccc",	
	753491: "cebdaababcaaaaabcccdecbcdabcdcbbbbbccddd",	
	110198: "aaaaaccccceeeeedbdbdaabcdcbbcdbbbbbccccc",	
	705932: "bbbbbbbbbbbbbbbbbbbbxxxxxxxxxxxxxxxxxxxx",	
	537282: "abcdeeeeeedcbaeedcbaabcdedcbabcdeedcbabc",	
	484470: "eeeeeeeeeeeeeeeeeeeeabcdeedcbaabcdeedcba",	
	422417: "abcdeeedcbabcdedcbaxxbcxeeeeeeeeeeeeeeee",	
	761336: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",	
	392411: "xxxxxxxxxxxxxxxxxxxxdedeeeeeeeeeeeeeeeee",	
	212971: "abcdeeeddedeeeexxeeexxxxxxxxxxxxxxxxxxxx",	
	783424: "abcdeedcbabcdedcbabcbbbbbdccccxddddxaaaa",	
	133168: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",	
	648037: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	872413: "abcaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	639863: "abcddaaaaabbbbbdddddbabcdabcdeccaaabbbbb",	
	363580: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	985335: "ceececdxcdcadddcddcbcebedcdeedcbabddcdee",	
	168030: "eeeeecccccbbbbbxxxxxcccccbbbbbaaaaabcccc",	
	235850: "abcdeaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	798399: "aaaaabbbbbcccccdddddaaaaaeeeeedddddbbbbb",	
	731579: "bbbbbbbaaabbababbbbbeeeeeeeeeedddddddddd",	
	760018: "aaaaabbbbbcccccdddddbbbbbcccccdddddccccc",	
	636975: "abbbcbbcbcbaaaaccdccaaaaabbbbbcdaaeabcbc",	
	658013: "cccccccccccccccccccccccccccccccccccccccc",	
	631788: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",	
	990337: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",	
	236273: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",	
	300532: "abcdedcdedcbabcdedcbabcdedcbabcdedcbabcd",	
	988443: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",	
	123800: "aaaaabbbbbccccccccccaaaaacccccdddddbcdcb",	
	624245: "aaaaabcdeababcdabcdeaabdcabcdcabcdcabbbb",	
	913602: "abcdeaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",	
	188319: "bbbbbbbbbbbbbbbbbbbbaaaaaaaaaaaaaaaaaaaa",	
}

func setupTestDB() (*gorm.DB, error) {
	testDBPath := "../internal/database/scan-eval-db.db"
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
	totalUnrecognized := 0 

	for studentID, expectedAnswers := range expectedAnswer_both_page {
		student, err := repository.GetStudentByRegistrationNumber(db, uint(studentID), 1) // testID = 1
		if err != nil {
			t.Errorf("Študent %d nebol nájdený: %v\n", studentID, err)
			totalMissing += len(expectedAnswers)
			continue
		}
		fmt.Printf("-----------------------\n")
		recognizedAnswers := student.Answers
		//ak nie je nic v DB
		if len(recognizedAnswers) == 0 {
			t.Errorf("Študent %d: chýbajúce odpovede\n", studentID)
			totalMissing += len(expectedAnswers)
			continue
		}

		correctCount := 0
		missingCount := 0
		unrecognized := 0
		totalQuestions += len(expectedAnswers)

		for i := 0; i < len(expectedAnswers); i++ {
			if i >= len(recognizedAnswers) {
				t.Errorf("Študent %d, otázka %d: chýbajúca odpoveď\n", studentID, i+1)
				missingCount++
				continue
			}

			if recognizedAnswers[i] == expectedAnswers[i] {
				correctCount++
			} else if recognizedAnswers[i] == '0' { // Check for unrecognized answer
				totalUnrecognized++
				unrecognized++
				missingCount++
			}else {
				unrecognized++
				fmt.Printf("Študent %d: Otázka č. %d, očakávané %s, rozpoznané %s\n", studentID, i+1, string(expectedAnswers[i]), string(recognizedAnswers[i]))
			}
			
		}

		totalCorrect += correctCount
		totalMissing += missingCount
		fmt.Printf("Študent %d: správne %d/40, nesprávne %d, chýbajúce %d\n", studentID, correctCount, unrecognized, missingCount)

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

		
		zeroCount := strings.Count(recognizedAnswers, "0")

		
		if zeroCount == 40 {
			missingPages += 2 
		}
		if zeroCount == 20 {
			missingPages++
		}
	}

	fmt.Printf("Celkový počet chýbajúcich strán: %d\n", missingPages)
}


