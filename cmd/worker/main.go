package main

import (
	"distgo/internal/dao/mq"
	"distgo/internal/dao/redis"
	"distgo/internal/logger"
	"distgo/internal/setting"
	"github.com/hibiken/asynq"
	"log"
)

const ConfigFilePath = "configs/distgo.yaml"

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
	if err := mq.InitServer(setting.Conf.AsynqConfig); err != nil {

	}

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(mq.TypeSendCompileJobs, mq.HandleCompileJobs)

	if err := mq.AsynqServer.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}

}
