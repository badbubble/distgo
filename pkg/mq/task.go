package mq

import (
	"context"
	"distgo/pkg/parser"
	"distgo/pkg/redis"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"os"
	"path/filepath"
)

const (
	TypeSendCompileJobs = "compile_job"
)

func NewCompileJob(job *parser.CompileJob) (*asynq.Task, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendCompileJobs, payload), nil
}

func HandleCompileJobs(ctx context.Context, t *asynq.Task) error {
	var job parser.CompileJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	for _, dep := range job.Dependencies {
		// check dependencies
		filename := fmt.Sprintf(filepath.Join(job.Path, dep))
		fmt.Println("DEP " + filename)
		// check local file
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			// file does not exist
			// try to get file from redis
			// download file
			err = redis.ReadFileFromRedis(filename)
			if err != nil {
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
