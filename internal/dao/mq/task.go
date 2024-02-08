package mq

import (
	"context"
	"distgo/internal/dao/redis"
	"distgo/internal/helper"
	"distgo/pkg/parser"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

var CheckGoTidy = make(map[string]interface{})

const (
	TypeSendCompileJobs = "compile_job"
)

func NewCompileGroupJob(job *parser.CompileGroup) (*asynq.Task, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		zap.L().Error("payload, err := json.Marshal(job) failed in NewCompileJob",
			zap.Any("Job", job),
			zap.Error(err),
		)
		return nil, err
	}
	return asynq.NewTask(TypeSendCompileGroup, payload), nil
}

func NewCompileJob(job *parser.CompileJob) (*asynq.Task, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		zap.L().Error("payload, err := json.Marshal(job) failed in NewCompileJob",
			zap.Any("Job", job),
			zap.Error(err),
		)
		return nil, err
	}
	return asynq.NewTask(TypeSendCompileJobs, payload), nil
}

func HandleCompileJobs(ctx context.Context, t *asynq.Task) error {
	var job parser.CompileJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		zap.L().Error("HandleCompileJobs json.Unmarshal failed",
			zap.Error(err),
		)
		return err
	}
	// go mod tidy
	if _, ok := CheckGoTidy[job.ProjectPath]; !ok {
		zap.L().Info("Execute go mod tidy", zap.String("Project", job.ProjectPath))
		_, err := helper.ExecuteCommand(fmt.Sprintf("cd %s && go mod tidy", job.ProjectPath))
		if err != nil {
			return err
		}
		CheckGoTidy[job.ProjectPath] = nil
	}

	//for _, dep := range job.Dependencies {
	//	// check dependencies
	//	filename := fmt.Sprintf(filepath.Join(job.Path, dep))
	//	// check if the file created in redis
	//	for {
	//		ok, _ := redis.CheckFileStatue(filename)
	//		if ok {
	//			zap.L().Info("find dependencies",
	//				zap.String("filename", filename),
	//			)
	//			break
	//		}
	//		time.Sleep(1 * time.Second)
	//		zap.L().Info("Cant find dependencies, try again",
	//			zap.String("filename", filename),
	//		)
	//	}
	//	// try to get file from redis
	//	if err := redis.ReadFileFromRedis(filename); err != nil {
	//		zap.L().Error("Failed to pull file from redis",
	//			zap.String("file", filename),
	//			zap.Error(err),
	//		)
	//		fmt.Printf("Failed to pull file from redis: %v\n", err)
	//		return err
	//	}
	//
	//}
	// compile
	parser.Compile([]*parser.CompileJob{&job})
	// send it to redis
	for _, au := range job.Autonomy {
		//err := redis.SaveFileToRedis(filepath.Join(job.Path, au))
		//if err != nil {
		//	fmt.Println("save file to redis error")
		//	return err
		//}
		err := redis.UpdateFileStatus(fmt.Sprintf("%s_%s", job.MD5, au))
		if err != nil {
			fmt.Println("update status failed")
			return err
		}

	}
	return nil
}
