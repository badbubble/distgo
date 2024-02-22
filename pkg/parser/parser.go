package parser

import (
	"distgo/internal/helper"
	"distgo/internal/snowflake"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const BuildCommand = "cd %s && go build -p 1 -n -a -work %s && rm -rf /tmp/go-build/ && mkdir -p /tmp/go-build/"
const ModCommand = "cd %s && go mod tidy"
const ARGMAX = 100000
const GCCPrefix = "gcc"
const CCPrefix = "cc"
const WorkPathPattern = "$WORK"
const WorkPath = "/tmp/go-build/%s"
const CreatBuildPath = "mkdir -p /tmp/go-build/%s"

var AutonomyPattern = regexp.MustCompile(`mkdir -p /tmp/go-build/[0-9]{15}/([^/]+)/`)
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
	BuildPath   string
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

func removeGCCAndReplaceWorkPath(commands []string, path string) []string {
	var result []string
	for _, c := range commands {
		if !(strings.HasPrefix(c, GCCPrefix) || strings.HasPrefix(c, CCPrefix)) {
			c = strings.ReplaceAll(c, WorkPathPattern, path)
			result = append(result, c)
		}
	}
	return result
}

func getGoBuildCommands(projectPath string, mainFile string) (string, error) {
	// go mod
	//command := fmt.Sprintf(ModCommand, projectPath)
	//_, err := helper.ExecuteCommand(command)
	//if err != nil {
	//	zap.L().Error("getGoBuildCommands generating building commands err",
	//		zap.String("ProjectPath", projectPath),
	//		zap.String("MainFile", mainFile),
	//		zap.String("Command", command),
	//		zap.Error(err),
	//	)
	//	return "", err
	//}

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
	outputSplit := strings.Split(output, "\n")
	var outputWithoutGet []string
	for _, o := range outputSplit {
		if !strings.HasPrefix(o, "# get ") {
			outputWithoutGet = append(outputWithoutGet, o)
		}
	}
	output = strings.Join(outputWithoutGet, "\n")
	return output, nil
}

func Compile(compileJobs []*CompileJob) {
	for _, job := range compileJobs {
		command := "cd " + job.ProjectPath + "\n" + "WORK=" + job.Path + "\n" + strings.Join(job.Commands, "\n")
		if len(command) < ARGMAX {
			cmd := exec.Command("sh", "-c", command)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("command exec err: %v, %s", err, string(output))
			}
		} else {
			tmpfile, err := os.Create("long_command")
			if err != nil {
				log.Fatalln(err)
			}
			// defer os.Remove(tmpfile.Name())
			if _, err := tmpfile.Write([]byte(command)); err != nil {
				log.Fatalln(err)
			}
			if err := tmpfile.Close(); err != nil {
				log.Fatalln(err)
			}
			cmd := exec.Command("sh", tmpfile.Name())
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("command exec err: %v, %s", err, string(output))
			}
		}

	}
}

func NewJobsByCommands(commandStr string, projectPath string, mainFile string) ([]*CompileGroup, error) {
	buildPath := snowflake.GenID()
	command := fmt.Sprintf(CreatBuildPath, buildPath)
	_, err := helper.ExecuteCommand(command)
	if err != nil {
		zap.L().Error("NewJobsByCommands create build path err",
			zap.String("ProjectPath", projectPath),
			zap.String("MainFile", mainFile),
			zap.String("Command", command),
			zap.Error(err),
		)
		return nil, err
	}
	var rawCommands = strings.Split(commandStr, "\n")
	// workDir := rawCommands[0]
	zap.L().Info("generate commands success",
		zap.String("ProjectPath", projectPath),
		zap.String("MainFile", mainFile),
		zap.Int("NumbersOfLine", len(rawCommands)),
	)
	var commands [][]string
	idx := 0
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

	var result []*CompileJob
	var gTasks []helper.Task
	var idToTask = make(map[string]*CompileJob)
	finalExecutionGroup := &CompileGroup{
		ID:          0,
		BuildPath:   buildPath,
		ProjectPath: "",
		Length:      0,
		Jobs:        []*CompileJob{},
	}
	var nCommands []string
	for _, c := range commands {
		for i, line := range c {
			// Check if the line ends with 'EOF' and 'EOF' is not the only text in the line
			if strings.HasSuffix(line, "EOF") && line != "EOF" {
				// Move 'EOF' to the new line
				// fmt.Println(line)
				c[i] = strings.TrimSuffix(line, "EOF") + "\nEOF"
			}
		}
		c = removeGCCAndReplaceWorkPath(c, fmt.Sprintf(WorkPath, buildPath))

		command := strings.Join(c, "\n")
		// add processed string
		nCommands = append(nCommands, command)

		autonomy := extractAutonomy(command)
		dep := extractDependencyPattern(command, autonomy)
		if len(autonomy) == 0 {
			//fmt.Println(autonomy)
			continue
		}
		if autonomy[0] != "b001" {
			idToTask[autonomy[0]] = &CompileJob{
				Autonomy:     autonomy,
				Dependencies: dep,
				Commands:     c,
				Path:         WorkPath,
				ProjectPath:  projectPath,
			}
		}

		if autonomy[0] != "b001" {
			gTasks = append(gTasks, helper.Task{
				ID:           autonomy[0],
				Dependencies: dep,
			})
		} else {
			if len(finalExecutionGroup.Jobs) == 0 {
				finalExecutionGroup.Jobs = append(finalExecutionGroup.Jobs, &CompileJob{
					Autonomy:     autonomy,
					Dependencies: dep,
					Commands:     c,
					Path:         WorkPath,
					ProjectPath:  projectPath,
				})
				finalExecutionGroup.Length = 1
			} else {
				finalExecutionGroup.Jobs[0].Commands = append(finalExecutionGroup.Jobs[0].Commands, c...)
			}

		}
	}
	if err := helper.WriteToFile("commands.sh", strings.Join(nCommands, "\n")); err != nil {
		zap.L().Error("getGoBuildCommands write commands to file failed",
			zap.String("ProjectPath", projectPath),
			zap.String("MainFile", mainFile),
			zap.Error(err),
		)
	}

	groups := helper.GroupTasks(gTasks)
	var compileGroups []*CompileGroup
	for i, group := range groups {
		newGroup := &CompileGroup{
			ID:          i,
			BuildPath:   buildPath,
			ProjectPath: projectPath,
			Length:      0,
			Jobs:        []*CompileJob{},
		}
		for _, id := range group {
			newGroup.Jobs = append(newGroup.Jobs, idToTask[id])
			newGroup.Length += 1
		}
		compileGroups = append(compileGroups, newGroup)
		//fmt.Printf("Group %d: %v\n", i+1, group)
	}
	finalExecutionGroup.ID = len(compileGroups)
	compileGroups = append(compileGroups, finalExecutionGroup)

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
