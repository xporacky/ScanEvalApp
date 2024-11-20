package files

import (
	"fmt"
	"os"
)

// OpenFile load file from specified path
func OpenFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("chyba pri otváraní súboru: %v", err)
	}
	return data, nil
}

// SaveFile saves []byte data to specified path
func SaveFile(filePath string, data []byte) error {
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("chyba pri ukladaní súboru: %v", err)
	}
	return nil
}

// DeleteFile deletes file from specified path
func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		// File exists, attempt to remove it
		err = os.Remove(filePath)
		if err != nil {
			fmt.Println("Error while deleting file:", err)
			return err
		}
		fmt.Println("Successfully deleted file:", filePath)
	}
	return nil
}
