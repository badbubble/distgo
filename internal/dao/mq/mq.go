package mq

import (
	"distgo/internal/setting"
	"fmt"
	"github.com/hibiken/asynq"
)

var AsynqClient *asynq.Client

func InitServer() error {
	return nil
}

func InitClient(config *setting.AsynqConfig) error {
	AsynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port)})
	return nil
}
