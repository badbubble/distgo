package mq

import (
	"context"
	"distgo/internal/dao/redis"
	"distgo/pkg/parser"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

const (
	TypeSendCompileJobs = "compile_job"
)

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

	for _, dep := range job.Dependencies {
		// check dependencies
		filename := fmt.Sprintf(filepath.Join(job.Path, dep))
		// check local file
		// file does not exist
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			zap.L().Info("dependencies required!",
				zap.String("filename", filename),
			)
			// try to get file from redis
			err = redis.ReadFileFromRedis(filename)
			if err != nil {
				zap.L().Error("Failed to pull file from redis",
					zap.String("file", filename),
					zap.Error(err),
				)
				fmt.Printf("Failed to pull file from redis: %v\n", err)
				return err
			}
		}

	}
	// compile
	parser.Compile([]*parser.CompileJob{&job})
	// send it to redis
	for _, au := range job.Autonomy {
		err := redis.SaveFileToRedis(filepath.Join(job.Path, au))
		if err != nil {
			fmt.Println("save file to redis error")
			return err
		}
	}
	return nil
}
