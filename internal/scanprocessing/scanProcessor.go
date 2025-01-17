package scanprocessing

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
	"gorm.io/gorm"
)

// Process PDF
func ProcessPDF(scanPath string, outputPath string, test *models.Test, db *gorm.DB) {
	var files []string

	err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".pdf" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		doc, err := fitz.New(file)
		if err != nil {
			panic(err)
		}
		//folder := strings.TrimSuffix(path.Base(file), filepath.Ext(path.Base(file)))

		// Extract pages as images
		for n := 0; n < doc.NumPage(); n++ {
			ProcessPage(doc, n, test, db)
		}
	}
}

func ProcessPage(doc *fitz.Document, n int, test *models.Test, db *gorm.DB) {
	img, err := doc.Image(n)
	if err != nil {
		panic(err)
	}
	mat := ImageToMat(img)
	mat = MatToGrayscale(mat)
	mat = FixImageRotation(mat)
	studentID, err := GetStudentID(&mat)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	student, err := repository.GetStudent(db, uint(studentID), test.ID)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("ID STUDENTA: ", studentID)
	//path := filepath.Join("./"+outputPath+"/", fmt.Sprintf("%s-image-%05d.png", folder, n)) //na testovanie zatial takto
	EvaluateAnswers(&mat, test.QuestionCount, student)
	err = repository.UpdateStudent(db, student)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	println(student.Answers)
	//SaveMat("", mat)
	defer mat.Close()
	println(student.Answers)
	ShowMat(mat)
	//return
}
