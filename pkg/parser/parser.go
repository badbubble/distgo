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
	MD5          string
	Autonomy     []string
	Dependencies []string
	Commands     []string
	Path         string
	ProjectPath  string
}

type CompileGroup struct {
	ID          int
	MD5         string
	ProjectPath string
	Length      int
	Jobs        []*CompileJob
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

func NewJobsByCommands(commandStr string, projectPath string, mainFile string) ([]*CompileGroup, error) {
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
	var gTasks []helper.Task
	var idToTask = make(map[string]*CompileJob)

	for _, c := range commands {
		for i, line := range c {
			// Check if the line ends with 'EOF' and 'EOF' is not the only text in the line
			if strings.HasSuffix(line, "EOF") && line != "EOF" {
				// Move 'EOF' to the new line
				fmt.Println(line)
				c[i] = strings.TrimSuffix(line, "EOF") + "\nEOF"
			}
		}
		command := strings.Join(c, "\n")
		autonomy := extractAutonomy(command)
		dep := extractDependencyPattern(command, autonomy)

		//result = append(result, &CompileJob{
		//	Autonomy:     autonomy,
		//	Dependencies: dep,
		//	Commands:     c,
		//	Path:         strings.Split(workDir, "=")[1],
		//	ProjectPath:  projectPath,
		//})
		idToTask[autonomy[0]] = &CompileJob{
			Autonomy:     autonomy,
			Dependencies: dep,
			Commands:     c,
			Path:         strings.Split(workDir, "=")[1],
			ProjectPath:  projectPath,
		}

		if autonomy[0] != "b001" || len(dep) != 0 {
			gTasks = append(gTasks, helper.Task{
				ID:           autonomy[0],
				Dependencies: dep,
			})
		}

	}
	//for _, task := range gTasks {
	//	fmt.Println(task)
	//}
	groups := helper.GroupTasks(gTasks)
	var compileGroups []*CompileGroup
	for i, group := range groups {
		newGroup := &CompileGroup{
			ID:          i,
			ProjectPath: projectPath,
			Length:      0,
			Jobs:        []*CompileJob{},
		}
		for _, id := range group {
			newGroup.Jobs = append(newGroup.Jobs, idToTask[id])
		}
		compileGroups = append(compileGroups, newGroup)
		fmt.Printf("Group %d: %v\n", i+1, group)
	}
	zap.L().Info("split commands success",
		zap.String("ProjectPath", projectPath),
		zap.String("MainFile", mainFile),
		zap.Int("NumbersOfCommands", len(result)),
	)
	return compileGroups, nil
}

func New(projectPath string, mainFile string) ([]*CompileGroup, error) {
	commandStr, err := getGoBuildCommands(projectPath, mainFile)
	if err != nil {
		return nil, err
	}
	return NewJobsByCommands(commandStr, projectPath, mainFile)
}
