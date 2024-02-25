package main

import (
	"distgo/internal/dao/mq"
	"distgo/internal/dao/redis"
	"distgo/internal/logger"
	"distgo/internal/setting"
	"distgo/internal/snowflake"
	"flag"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
	"os"
	"time"
)

func main() {
	exitChan := make(chan bool)

	// Start a new goroutine
	go func() {
		startTime := time.Now()
		for {
			if _, err := os.Stat("/home/ppp23114/Projects/Projects/wego/main"); err == nil {
				cost := time.Since(startTime).Seconds()
				fmt.Println("############### Cost Time ###############")
				fmt.Println(cost)
				fmt.Println("####################################################")
				exitChan <- true
				return
			}
		}
	}()

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
	// init snowflake
	if err := snowflake.Init(); err != nil {
		log.Fatalf("init snowflake failed: %v", err)
	}
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(mq.TypeSendCompileGroup, mq.HandleCompileGroup)

	if err := mq.AsynqServer.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}

	// Wait for the signal to exit
	<-exitChan
}
