package latex

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func CompileLatexToPDF(latexContent []byte) ([]byte, error) {
	// Create a temporary .tex file
	texFile, err := ioutil.TempFile("", "*.tex")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary LaTeX file: %v", err)
	}
	defer os.Remove(texFile.Name())

	// Write LaTeX content to the temporary file
	if _, err = texFile.Write(latexContent); err != nil {
		return nil, fmt.Errorf("error writing to .tex file: %v", err)
	}
	texFile.Close()

	// Create a temporary directory for the output
	outputDir, err := ioutil.TempDir("", "latex_output")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary output directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Run pdflatex to compile
	cmd := exec.Command("pdflatex", "-output-directory", outputDir, texFile.Name())
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("error compiling LaTeX to PDF: %v", err)
	}

	// Read the PDF file as bytes
	pdfPath := filepath.Join(outputDir, filepath.Base(texFile.Name())[:len(filepath.Base(texFile.Name()))-4]+".pdf")
	pdfBytes, err := ioutil.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("error reading PDF file: %v", err)
	}

	return pdfBytes, nil
}

// LaTeX pdf generation

// Funkcia na nahradenie place holderov v LaTeX súbore
func ReplaceTemplatePlaceholders(templateContent []byte, data TemplateData) ([]byte, error) {
	tmpl, err := template.New("latex").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("chyba pri parsovaní šablóny: %v", err)
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		return nil, fmt.Errorf("chyba pri nahrádzaní hodnôt v šablóne: %v", err)
	}

	return output.Bytes(), nil
}
