package main

import (
	"fmt"
	"os"
	"os/exec"
)

func CompileLatexToPDF(latexFilePath string) error {
	if _, err := os.Stat(latexFilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", latexFilePath)
	}

	cmd := exec.Command("pdflatex", latexFilePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to compile LaTeX file: %v", err)
	}

	return nil
}

func main() {
	latexFilePath := "./latexFiles/main.tex"
	err := CompileLatexToPDF(latexFilePath)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("PDF compiled successfully.")
	}
}
