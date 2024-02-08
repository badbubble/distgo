package main

import (
	"distgo/internal/dao/mq"
	"distgo/internal/dao/redis"
	"distgo/internal/logger"
	"distgo/internal/setting"
	"flag"
	"github.com/hibiken/asynq"
	"log"
)

func main() {
	var ConfigFilePath string

	flag.StringVar(&ConfigFilePath, "c", "configs/coordinator.yaml", "the file path of the master.yaml")
	flag.Parse()

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
	if err := mq.InitServerCoordinator(setting.Conf.AsynqConfig); err != nil {

	}
	if err := mq.InitClient(setting.Conf.AsynqConfig); err != nil {
		log.Fatalf("create asynq client failed: %v", err)
	}

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(mq.TypeSendCompileGroup, mq.HandleCompileGroup)

	if err := mq.AsynqServer.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}