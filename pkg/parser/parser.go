package parser

import (
	"distgo/internal/helper"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
	"regexp"
	"strings"
)

const BuildCommand = "cd %s && go build -p 1 -x -a -work %s && rm main"

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

func getGoBuildCommands(projectPath string, mainFile string) (string, error) {
	// generate the go commands
	command := fmt.Sprintf(BuildCommand, projectPath, mainFile)
	output, err := helper.ExecuteCommand(command)
	if err != nil {
		zap.L().Error("getGoBuildCommands generating building commands err",
			zap.String("ProjectPath", projectPath),
			zap.String("MainFile", mainFile),
			zap.String("Command", command),
			zap.Error(err),
		)
		return "", err
	}
	if err := helper.WriteToFile("commands.sh", output); err != nil {
		zap.L().Error("getGoBuildCommands write commands to file failed",
			zap.String("ProjectPath", projectPath),
			zap.String("MainFile", mainFile),
			zap.Error(err),
		)
	}
	return output, nil
}

func Compile(compileJobs []*CompileJob) {
	for _, job := range compileJobs {
		command := "cd " + job.ProjectPath + "\n" + "WORK=" + job.Path + "\n" + strings.Join(job.Commands, "\n")
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("command exec err: %v, %s", err, string(output))
		}

	}
}

func NewJobsByCommands(commandStr string, projectPath string, mainFile string) ([]*CompileJob, error) {
	var rawCommands = strings.Split(commandStr, "\n")
	workDir := rawCommands[0]
	zap.L().Info("generate commands success",
		zap.String("ProjectPath", projectPath),
		zap.String("MainFile", mainFile),
		zap.Int("NumbersOfLine", len(rawCommands)),
	)
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

	//var IndependentCommands [][]string
	//
	//joinFlag := false
	//for _, c := range commands {
	//	if joinFlag {
	//		IndependentCommands[len(IndependentCommands)-1] = append(IndependentCommands[len(IndependentCommands)-1], c...)
	//	} else {
	//		IndependentCommands = append(IndependentCommands, c)
	//	}
	//
	//	command := strings.Join(c, "\n")
	//	if strings.Contains(command, "cd /") {
	//		joinFlag = true
	//	} else {
	//		joinFlag = false
	//	}
	//}
	var result []*CompileJob

	for _, c := range commands {
		command := strings.Join(c, "\n")
		autonomy := extractAutonomy(command)
		dep := extractDependencyPattern(command, autonomy)
		result = append(result, &CompileJob{
			Autonomy:     autonomy,
			Dependencies: dep,
			Commands:     c,
			Path:         strings.Split(workDir, "=")[1],
			ProjectPath:  projectPath,
		})
	}
	zap.L().Info("split commands success",
		zap.String("ProjectPath", projectPath),
		zap.String("MainFile", mainFile),
		zap.Int("NumbersOfCommands", len(result)),
	)
	return result, nil
}

func New(projectPath string, mainFile string) ([]*CompileJob, error) {
	commandStr, err := getGoBuildCommands(projectPath, mainFile)
	if err != nil {
		return nil, err
	}
	return NewJobsByCommands(commandStr, projectPath, mainFile)
}
