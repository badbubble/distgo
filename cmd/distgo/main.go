package main

import (
	"distgo/internal/dao/mq"
	"distgo/internal/dao/redis"
	"distgo/internal/logger"
	"distgo/internal/setting"
	"distgo/pkg/parser"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"log"
	"time"
)

const ConfigFilePath = "configs/distgo.yaml"
const ProjectPath = "/Users/badbubble/GolandProjects/BubblePL"
const MainFile = "main.go"

func main() {
	// read configurations
	if err := setting.Init(ConfigFilePath); err != nil {
		log.Fatalf("read configurations failed: %v", err)
	}
	// initiate logger
	if err := logger.Init(setting.Conf.LogConfig, setting.Conf.Mode); err != nil {
		log.Fatalf("initiate logger failed: %v", err)
	}
	// connect to redis
	if err := redis.Init(setting.Conf.RedisConfig); err != nil {
		log.Fatalf("connect to client failed: %v", err)
	}
	// create asynq
	if err := mq.InitClient(setting.Conf.AsynqConfig); err != nil {
		log.Fatalf("create asynq client failed: %v", err)
	}

	//split the go build commands to compileJobs
	compileJobs, err := parser.New(ProjectPath, MainFile)
	if err != nil {
		zap.L().Error("parser.New failed",
			zap.String("ProjectPath", ProjectPath),
			zap.String("MainFile", MainFile),
			zap.Error(err),
		)
		return
	}
	// send the compileJobs to Asynq
	for _, job := range compileJobs {
		task, err := mq.NewCompileJob(job)
		if err != nil {
			zap.L().Error("failed to create new job ",
				zap.Any("Job", job),
				zap.Error(err),
			)
			return
		}
		if _, err = mq.AsynqClient.Enqueue(task, asynq.Retention(24*time.Hour), asynq.MaxRetry(10)); err != nil {
			zap.L().Error("failed to put task in asynq",
				zap.Any("task", task),
				zap.Error(err),
			)
		}
	}
}
