package ocr

import "os/exec"

const PSM_SINGLE_LINE = "7"
const PSM_DEFAULT = "3"

func OcrImage(imagePath string, psm string) string {

	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "slk", "--psm", psm)
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}
