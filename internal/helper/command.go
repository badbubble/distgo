package helper

import (
	"os/exec"
)

// ExecuteCommand executes bash command using exec.Command and returns the output and error
func ExecuteCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
