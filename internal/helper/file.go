package helper

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"
	"strings"
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

func CheckEOFError(commands []string) []string {

	// Process each line
	for i, line := range commands {
		// Check if the line ends with 'EOF' and 'EOF' is not the only text in the line
		if strings.HasSuffix(line, "EOF") && line != "EOF" {
			// Move 'EOF' to the new line
			commands[i] = strings.TrimSuffix(line, "EOF") + "\nEOF"
		}
	}
	return commands
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
