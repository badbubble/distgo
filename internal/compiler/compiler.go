package compiler

import (
	"distgo/internal/dao/redis"
	"distgo/pkg/parser"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

func Compile(job *parser.CompileJob) error {
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
			for errors.Is(err, redis.ErrNoKeyFound) {
				zap.L().Info("Try to get the file again",
					zap.String("filename", filename),
				)
				err = redis.ReadFileFromRedis(filename)
				time.Sleep(1 * time.Second)
			}

			if err != nil {
				zap.L().Error("Failed to pull file from redis",
					zap.String("file", filename),
					zap.Error(err),
				)
				return err
			}
		}

	}
	// compile
	parser.Compile([]*parser.CompileJob{job})
	// send it to redis
	for _, au := range job.Autonomy {
		err := redis.SaveFileToRedis(filepath.Join(job.Path, au))
		if err != nil {
			fmt.Println("save file to redis err")
			return err
		}
	}
	return nil
}
