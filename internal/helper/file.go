package helper

import (
	"log"
	"os"
)

func DeleteFile() {

}

func WriteToFile(filePath string, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the string to the file
	if _, err = file.WriteString(content); err != nil {
		return err
	}
	return nil
}

func ReadFromFile(filepath string) (string, error) {
	body, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
		return "", err
	}
	return string(body), err
}
