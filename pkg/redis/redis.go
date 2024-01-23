package redis

import (
	"context"
	"distgo/internal/setting"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"path/filepath"
)

var (
	client *redis.Client
	Nil    = redis.Nil
)
var ctx = context.Background()

func SaveFileToRedis(filePath string) error {
	root := fmt.Sprintf("%s", filePath)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			fmt.Println()
			return err
		}
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			data, err := os.ReadFile(path)

			err = client.Set(context.Background(), path, data, 0).Err()
			if err != nil {
				return err
			}
		} else {
			err = client.Set(context.Background(), path, "IT IS A PATH", 0).Err()
			if err != nil {
				return err
			}
		}

		return nil
	})

	return nil
}

func ReadFileFromRedis(filePath string) error {
	fmt.Printf("TRY TO GET FIEL")
	pattern := fmt.Sprintf("%s/*", filePath)
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		fmt.Println("Error getting keys:", err)
		return err
	}

	// Iterating over keys to get their values
	for _, key := range keys {
		value, err := client.Get(ctx, key).Result()
		if err != nil {
			fmt.Println("Error getting value for key", key, ":", err)
			continue
		}

		// Create the directory structure for the file if it doesn't exist
		dir := filepath.Dir(key)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Println("Error creating directory for key", key, ":", err)
			continue
		}
		if value != "IT IS A PATH" {
			// Write the value to a file
			if err := os.WriteFile(key, []byte(value), 0644); err != nil {
				fmt.Println("Error writing file for key", key, ":", err)
				continue
			}
		}
	}
	return nil
}

// Init
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
