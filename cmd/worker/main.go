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

	flag.StringVar(&ConfigFilePath, "c", "configs/worker.yaml", "the file path of the master.yaml")
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
	if err := mq.InitServer(setting.Conf.AsynqConfig); err != nil {

	}

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(mq.TypeSendCompileJobs, mq.HandleCompileJobs)

	if err := mq.AsynqServer.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}

	//var stop string
	//for {
	//	fmt.Println(">")
	//	fmt.Scanln(&stop)
	//	fmt.Println("EXECUTE")
	//	result, _ := helper.ReadFromFile("commands_s.sh")
	//	jobs, _ := parser.NewJobsByCommands(result, "/Users/badbubble/GolandProjects/BubblePL", "main.go")
	//	zap.L().Info("the job is",
	//		zap.Any("dep", jobs[0].Dependencies),
	//		zap.Int("Length", len(jobs)),
	//	)
	//	compiler.Compile(jobs[0])
	//	fmt.Println("DONE")
	//}

}
