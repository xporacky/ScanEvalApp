package main

//SPUSTANIE V PRIECINKU:
// go run cmd/param-eval/main.go -pdf cesta/k/suboru.pdf -choices 5 -questions 20 -config Zasadacka_200dpi
import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ScanEvalApp/internal/logging"
	"ScanEvalApp/internal/scanprocessing"

	"github.com/gen2brain/go-fitz"
)

func main() {
	var (
		pdfPath   = flag.String("pdf", "", "Cesta k PDF so skenmi hárkov (vyžadované)")
		choices   = flag.Int("choices", 5, "Počet možností na otázku (napr. 5)")
		questions = flag.Int("questions", 0, "Počet otázok na hárku (vyžadované)")
		config    = flag.String("config", "Zasadacka_200dpi", "Názov konfigurácie (bez .json) z configs/")
	)
	flag.Parse()

	if *pdfPath == "" {
		exitWithError(errors.New("chýba parameter -pdf"))
	}
	if *questions <= 0 {
		exitWithError(errors.New("chýba alebo je neplatný parameter -questions (> 0)"))
	}

	absPDF, err := filepath.Abs(*pdfPath)
	if err != nil {
		exitWithError(fmt.Errorf("neviem získať absolútnu cestu: %w", err))
	}
	if _, err := os.Stat(absPDF); err != nil {
		exitWithError(fmt.Errorf("súbor neexistuje alebo je nedostupný: %w", err))
	}

	logging.InitLogger()
	if err := scanprocessing.LoadConfig(*config); err != nil {
		exitWithError(fmt.Errorf("nepodarilo sa načítať config %q: %w", *config, err))
	}

	doc, err := fitz.New(absPDF)
	if err != nil {
		exitWithError(fmt.Errorf("chyba pri otváraní PDF: %w", err))
	}
	defer doc.Close()

	total := doc.NumPage()
	fmt.Printf("Spracúvam %d strán(y) z %s\n", total, absPDF)

	for i := 0; i < total; i++ {
		img, err := doc.Image(i)
		if err != nil {
			fmt.Printf("Strana %d: chyba extrakcie obrázka: %v\n", i+1, err)
			continue
		}
		answers, err := scanprocessing.DetectAnswersFromImage(img, *choices, *questions)
		if err != nil {
			fmt.Printf("Strana %d: chyba pri detekcii: %v\n", i+1, err)
			continue
		}
		fmt.Printf("Strana %d: %s\n", i+1, string(answers))
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
