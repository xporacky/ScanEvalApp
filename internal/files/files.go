package files

import (
	"fmt"
	"os"
)

// OpenFile načíta obsah súboru a vráti ho ako []byte
func OpenFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("chyba pri otváraní súboru: %v", err)
	}
	return data, nil
}

// SaveFile uloží obsah []byte do súboru na špecifikovanej ceste
func SaveFile(filePath string, data []byte) error {
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("chyba pri ukladaní súboru: %v", err)
	}
	return nil
}
