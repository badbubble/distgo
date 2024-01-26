package redis

import (
	"context"
	"distgo/internal/setting"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

// RootPathContent root path for example: b100, b101
var RootPathContent = "IT IS A PATH"

var (
	client *redis.Client
	Nil    = redis.Nil
)

var ctx = context.Background()

// SaveFileToRedis writes all files under filePath to Redis
func SaveFileToRedis(filePath string) error {
	zap.L().Info("Writing files to Redis",
		zap.String("Path", filePath),
	)
	root := fmt.Sprintf("%s", filePath)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			zap.L().Error("Failed to find files under filePath",
				zap.String("Path", filePath),
				zap.Error(err),
			)
			return err
		}
		zap.L().Info("Found a file ready to write to Redis",
			zap.String("Path", path),
		)
		fileInfo, err := os.Stat(path)
		if err != nil {
			zap.L().Error("Failed to get stat from a file",
				zap.String("Path", path),
				zap.Error(err),
			)
			return err
		}
		if !fileInfo.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				zap.L().Error("Failed to read the file",
					zap.String("Path", path),
					zap.Error(err),
				)
				return err
			}
			zap.L().Info("File Read",
				zap.String("Path", path),
				zap.Int("Length", len(data)),
			)

			err = client.Set(context.Background(), path, data, 0).Err()
			if err != nil {
				zap.L().Error("Failed to write data to Redis",
					zap.String("Path", path),
					zap.Int("Length", len(data)),
				)
				return err
			}
		} else {
			err = client.Set(context.Background(), path, RootPathContent, 0).Err()
			if err != nil {
				zap.L().Error("Failed to create data to Redis",
					zap.String("Path", path),
					zap.Error(err),
				)
				return err
			}
			zap.L().Info("Root Path created",
				zap.String("Path", path),
			)

		}

		return nil
	})

	return nil
}

// ReadFileFromRedis can get all files from filePath for example
func ReadFileFromRedis(filePath string) error {
	pattern := fmt.Sprintf("%s/*", filePath)
	zap.L().Info("Try to get file from Redis",
		zap.String("FilePath", filePath),
		zap.String("Pattern", pattern),
	)

	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		zap.L().Error("Failed to get file from Redis by pattern",
			zap.String("Pattern", pattern),
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("found keys in Redis",
		zap.Any("keys", keys),
		zap.Int("Length", len(keys)),
	)

	// Iterating over keys to get their values
	for _, key := range keys {
		value, err := client.Get(ctx, key).Result()
		if err != nil {
			zap.L().Error("Failed to get file from Redis by key",
				zap.String("Key", key),
				zap.Error(err),
			)
			return err
		}
		zap.L().Info("get file success by key",
			zap.String("key", key),
			zap.Int("Length", len(value)),
		)
		// Create the directory structure for the file if it doesn't exist
		dir := filepath.Dir(key)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			zap.L().Error("Failed to create file",
				zap.String("Path", key),
				zap.Error(err),
			)
			return err
		}

		zap.L().Info("created a path or file",
			zap.String("path", key),
		)

		if value != RootPathContent {
			// Write the value to a file
			if err := os.WriteFile(key, []byte(value), 0644); err != nil {
				zap.L().Error("Failed to write to a file",
					zap.String("Path", key),
					zap.Error(err),
				)
				return err
			}
			zap.L().Info("write to a file success",
				zap.String("path", key),
			)
		}
	}
	return nil
}

// Init Redis to get a redis client
func Init(cfg *setting.RedisConfig) (err error) {
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password, // no password set
		DB:           cfg.DB,       // use default DB
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}

func Close() {
	_ = client.Close()
}
