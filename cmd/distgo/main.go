package main

import (
	"distgo/internal/setting"
	"distgo/pkg/mq"
	"distgo/pkg/parser"
	"distgo/pkg/redis"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
	"os"
	"os/exec"
	"time"
)

const ConfigFilePath = "configs/distgo.yaml"

func getGoBuildCommands(projectPath string) string {
	command := fmt.Sprintf("cd %s && go build -x -a -work main.go", projectPath)
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("generate command error: %v", err)
	}
	file, err := os.Create("commands.sh")
	if err != nil {
		fmt.Println("Error creating the file:", err)

	}
	defer file.Close()

	// Write the string to the file
	_, err = file.WriteString(string(output))
	if err != nil {
		fmt.Println("Error writing to the file:", err)
	}
	return string(output)
}

func main() {
	err := setting.Init("/home/badbubble/distgo/configs/distgo.yaml")
	if err != nil {
		log.Fatalf("read settings failed: %v", err)
	}
	err = redis.Init(setting.Conf.RedisConfig)
	if err != nil {
		log.Fatalf("create redis client failed: %v", err)
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", setting.Conf.AsynqConfig.Host, setting.Conf.RedisConfig.Port)})

	// generate the go build commands
	projectPath := "/home/badbubble/BubblePL"
	commands := getGoBuildCommands("/home/badbubble/BubblePL")
	//split the go build commands to compileJobs
	compileJobs, err := parser.New(commands, projectPath)
	if err != nil {
		log.Fatalf("create compile jobs error: %v", compileJobs)
	}
	// send the compileJobs to asynq

	for _, job := range compileJobs {
		task, err := mq.NewCompileJob(job)
		if err != nil {
			log.Fatalf("send to asynq failed: %v", err)
		}
		_, err = client.Enqueue(task, asynq.Retention(24*time.Hour), asynq.MaxRetry(10), asynq.Timeout(3*time.Minute))
		if err != nil {
			log.Fatalf("could not enqueue task: %v", err)
		}
	}

	// collect the complieJobs Result from asynq
	// finish the compiling

}
