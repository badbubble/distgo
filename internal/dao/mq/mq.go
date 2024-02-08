package mq

import (
	"distgo/internal/setting"
	"fmt"
	"github.com/hibiken/asynq"
)

var AsynqClient *asynq.Client
var AsynqServer *asynq.Server

func InitServerWorker(config *setting.AsynqConfig) error {
	AsynqServer = asynq.NewServer(
		asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port)},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: config.Concurrency,
			Queues: map[string]int{
				"Compile_Job": 1,
			},
		},
	)
	return nil
}

func InitServerCoordinator(config *setting.AsynqConfig) error {
	AsynqServer = asynq.NewServer(
		asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port)},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: config.Concurrency,
			Queues: map[string]int{
				"Compile_Group": 1,
			},
		},
	)
	return nil
}
func InitClient(config *setting.AsynqConfig) error {
	AsynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port)})
	return nil
}
