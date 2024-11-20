package ocr

import "os/exec"

func OcrImage(imagePath string) string {

	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "slk", "--psm", "3")
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}
