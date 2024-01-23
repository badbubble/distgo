package parser

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var AutonomyPattern = regexp.MustCompile(`mkdir -p \$WORK/([^/]+)/`)
var DependencyPattern = regexp.MustCompile(`/b\d{3}/`)

type CompileJob struct {
	Autonomy     []string
	Dependencies []string
	Commands     []string
	Path         string
	ProjectPath  string
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func extractAutonomy(command string) []string {
	var result []string
	matches := AutonomyPattern.FindAllStringSubmatch(command, -1)
	for _, match := range matches {
		if len(match) > 1 {
			result = append(result, match[1])
		}
	}
	return result
}

func extractDependencyPattern(command string, autonomy []string) []string {
	var result []string
	matches := DependencyPattern.FindAllStringSubmatch(command, -1)
	setResult := map[string]struct{}{}

	for _, match := range matches {
		if len(match) == 1 {
			work := strings.Trim(match[0], "/")
			if !stringInSlice(work, autonomy) {
				setResult[work] = struct{}{}
			}
		}
	}
	for k, _ := range setResult {
		result = append(result, k)
	}
	return result
}

func Compile(compileJobs []*CompileJob) {
	for _, job := range compileJobs {
		command := "cd " + job.ProjectPath + "\n" + "WORK=" + job.Path + "\n" + strings.Join(job.Commands, "\n")
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("command exec error: %v, %s", err, string(output))
		}

		//fmt.Println(command)
		//fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	}
}

func New(commandStr string, projectPath string) ([]*CompileJob, error) {
	var rawCommands = strings.Split(commandStr, "\n")
	workDir := rawCommands[0]
	fmt.Println(len(rawCommands))
	var commands [][]string
	idx := 1
	for idx < len(rawCommands) {
		rc := rawCommands[idx]
		var linkCommand []string
		if strings.HasPrefix(rc, "mkdir") {
			linkCommand = append(linkCommand, rc)
			idx += 1
			if idx == len(rawCommands) {
				break
			}

			for strings.HasPrefix(rawCommands[idx], "mkdir") == false {
				linkCommand = append(linkCommand, rawCommands[idx])
				idx += 1
				if idx == len(rawCommands) {
					break
				}
			}
		} else {
			linkCommand = append(linkCommand, rc)
			idx += 1
		}

		commands = append(commands, linkCommand)
	}

	var IndependentCommands [][]string

	joinFlag := false
	for _, c := range commands {
		if joinFlag {
			IndependentCommands[len(IndependentCommands)-1] = append(IndependentCommands[len(IndependentCommands)-1], c...)
		} else {
			IndependentCommands = append(IndependentCommands, c)
		}

		command := strings.Join(c, "\n")
		if strings.Contains(command, "cd /") {
			joinFlag = true
		} else {
			joinFlag = false
		}
	}
	var result []*CompileJob

	for _, commands := range IndependentCommands {
		command := strings.Join(commands, "\n")
		autonomy := extractAutonomy(command)
		dep := extractDependencyPattern(command, autonomy)
		result = append(result, &CompileJob{
			Autonomy:     autonomy,
			Dependencies: dep,
			Commands:     commands,
			Path:         strings.Split(workDir, "=")[1],
			ProjectPath:  projectPath,
		})
	}
	return result, nil
}
