package helper

import (
	"fmt"
	"os/exec"
)

// ExecuteCommand executes bash command using exec.Command and returns the output and error
func ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func SendFilesTo(filePath string, host string) (string, error) {
	command := fmt.Sprintf("rsync -avz %s %s:%s", filePath, host, filePath)
	fmt.Println(command)
	return ExecuteCommand(command)
}

func RecvFilesFrom(filePath string, host string) (string, error) {
	command := fmt.Sprintf("rsync -avz %s:%s %s", host, filePath, filePath)
	fmt.Println(command)
	return ExecuteCommand(command)
}
