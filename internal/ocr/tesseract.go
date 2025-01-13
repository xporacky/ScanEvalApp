package ocr

import (
	"fmt"
	"os/exec"
	"regexp"
)

const PSM_SINGLE_LINE = "7"
const PSM_UNIFORM_BLOCK = "6"
const PSM_DEFAULT = "3"

func OcrImage(imagePath string, psm string) string {

	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "slk", "--psm", psm)
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}

func ExtractID(path string) (int, error) {
	dt := OcrImage(path, PSM_UNIFORM_BLOCK)
	re := regexp.MustCompile(`ID:\s*(\d+)`)
	match := re.FindStringSubmatch(dt)
	if len(match) < 2 {
		return 0, fmt.Errorf("no ID found in the input image")
	}
	var id int
	_, err := fmt.Sscan(match[1], &id)
	if err != nil {
		return 0, fmt.Errorf("failed to convert ID to integer: %v", err)
	}
	return id, nil
}

func ExtractQuestionNumber(path string) (int, error) {
	dt := OcrImage(path, PSM_SINGLE_LINE)
	var num int
	_, err := fmt.Sscan(dt, &num)
	if err != nil {
		return 0, fmt.Errorf("failed to convert QuestionNumber to integer: %v", err)
	}
	fmt.Println("Question number:", num)
	return num, nil
}
