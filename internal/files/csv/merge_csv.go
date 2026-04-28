package csv

import (
	"ScanEvalApp/internal/logging"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// MergeResultCSVs zlúči viaceré výsledkové CSV súbory do jedného.
// Stĺpce Podskupina, Skóre, Odpovede študenta, Správne odpovede sa prepisujú
// len vtedy, keď zdrojový riadok obsahuje neprázdnu hodnotu v stĺpci Odpovede študenta
// (t.j. nie samé nuly alebo prázdny reťazec).
// Kľúčom pre zlučovanie je Reg. číslo (stĺpec index 2).
// outputPath je úplná cesta výsledného súboru vrátane názvu.
func MergeResultCSVs(inputPaths []string, outputPath string) error {
	if len(inputPaths) == 0 {
		return fmt.Errorf("žiadne vstupné CSV súbory")
	}

	errorLogger := logging.GetErrorLogger()

	// Stĺpce ktoré sa zlučujú (indexy v 10-stĺpcovom formáte):
	// 0 Meno, 1 Priezvisko, 2 Reg. číslo, 3 Dátum skúšky, 4 Čas,
	// 5 Miestnosť, 6 Podskupina, 7 Skóre, 8 Odpovede študenta, 9 Správne odpovede
	const (
		colRegNum    = 2
		colSubgroup  = 6
		colScore     = 7
		colAnswers   = 8
		colCorrect   = 9
		expectedCols = 10
	)

	type row struct {
		data  []string
		order int // poradie z prvého súboru
	}

	// base: základ zo prvého súboru (zachová poradie riadkov)
	base := make(map[string]*row) // kľúč = Reg. číslo
	var keyOrder []string

	// Načítaj prvý súbor ako základ
	firstFile, err := os.Open(inputPaths[0])
	if err != nil {
		return fmt.Errorf("chyba pri otváraní súboru %s: %w", inputPaths[0], err)
	}
	firstReader := csv.NewReader(firstFile)
	firstReader.FieldsPerRecord = -1
	firstRows, err := firstReader.ReadAll()
	firstFile.Close()
	if err != nil {
		return fmt.Errorf("chyba pri čítaní súboru %s: %w", inputPaths[0], err)
	}
	if len(firstRows) == 0 {
		return fmt.Errorf("súbor %s je prázdny", inputPaths[0])
	}

	header := firstRows[0]
	for i, dataRow := range firstRows[1:] {
		if len(dataRow) < expectedCols {
			for len(dataRow) < expectedCols {
				dataRow = append(dataRow, "")
			}
		}
		regNum := strings.TrimSpace(dataRow[colRegNum])
		if regNum == "" {
			continue
		}
		copied := make([]string, expectedCols)
		copy(copied, dataRow[:expectedCols])
		base[regNum] = &row{data: copied, order: i}
		keyOrder = append(keyOrder, regNum)
	}

	// isFilledAnswers vráti true ak reťazec odpovedí obsahuje aspoň jednu
	// odpoveď inú ako '0' (t.j. naozaj bola spracovaná)
	isFilledAnswers := func(ans string) bool {
		ans = strings.TrimSpace(ans)
		if ans == "" {
			return false
		}
		for _, ch := range ans {
			if ch != '0' {
				return true
			}
		}
		return false
	}

	// Aplikuj každý ďalší súbor
	for _, path := range inputPaths[1:] {
		f, err := os.Open(path)
		if err != nil {
			errorLogger.Error("Chyba pri otváraní CSV na merge", slog.String("path", path), slog.String("error", err.Error()))
			return fmt.Errorf("chyba pri otváraní súboru %s: %w", path, err)
		}
		r := csv.NewReader(f)
		r.FieldsPerRecord = -1
		rows, err := r.ReadAll()
		f.Close()
		if err != nil {
			errorLogger.Error("Chyba pri čítaní CSV na merge", slog.String("path", path), slog.String("error", err.Error()))
			return fmt.Errorf("chyba pri čítaní súboru %s: %w", path, err)
		}
		for i, dataRow := range rows {
			if i == 0 {
				continue // preskoc hlavičku
			}
			if len(dataRow) < expectedCols {
				for len(dataRow) < expectedCols {
					dataRow = append(dataRow, "")
				}
			}
			regNum := strings.TrimSpace(dataRow[colRegNum])
			if regNum == "" {
				continue
			}
			src := dataRow
			// Zlúč iba ak tento súbor má vyplnené odpovede pre daného študenta
			if !isFilledAnswers(src[colAnswers]) {
				continue
			}
			if existing, ok := base[regNum]; ok {
				existing.data[colSubgroup] = src[colSubgroup]
				existing.data[colScore] = src[colScore]
				existing.data[colAnswers] = src[colAnswers]
				existing.data[colCorrect] = src[colCorrect]
			}
			// Ak reg. číslo nie je v základe, preskočíme – základ určuje štruktúru
		}
	}

	// Zapis výsledného súboru
	outFile, err := os.Create(outputPath)
	if err != nil {
		errorLogger.Error("Chyba pri vytváraní výstupného CSV", slog.String("path", outputPath), slog.String("error", err.Error()))
		return fmt.Errorf("chyba pri vytváraní výstupného súboru: %w", err)
	}
	defer outFile.Close()

	// UTF-8 BOM pre Excel
	outFile.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("chyba pri zápise hlavičky: %w", err)
	}

	for _, regNum := range keyOrder {
		r := base[regNum]
		if err := writer.Write(r.data); err != nil {
			errorLogger.Error("Chyba pri zápise riadku do výsledného CSV", slog.String("regNum", regNum), slog.String("error", err.Error()))
			return fmt.Errorf("chyba pri zápise riadku: %w", err)
		}
	}

	return nil
}
