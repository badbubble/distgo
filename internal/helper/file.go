package helper

import (
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
