package main

import (
	"distgo/internal/setting"
	"distgo/pkg/mq"
	"distgo/pkg/redis"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
)

func main() {
	err := setting.Init("/home/badbubble/distgo/configs/distgo.yaml")
	if err != nil {
		log.Fatalf("read settings failed: %v", err)
	}

	err = redis.Init(setting.Conf.RedisConfig)
	if err != nil {
		log.Fatalf("create redis client failed: %v", err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", setting.Conf.AsynqConfig.Host, setting.Conf.RedisConfig.Port)},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 2,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(mq.TypeSendCompileJobs, mq.HandleCompileJobs)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
