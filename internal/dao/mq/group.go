package mq

import (
	"context"
	"distgo/internal/dao/redis"
	"distgo/internal/helper"
	"distgo/internal/setting"
	"distgo/pkg/parser"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"sync"
	"time"
)

var JobDistributedTime = 0.0
var ExchangeDepTime = 0.0

const TypeSendCompileGroup = "compile_group"

func SendProjectsToWorker() error {
	for _, host := range setting.Conf.ClusterConfig.Hosts {
		_, err := helper.SendFilesTo(setting.Conf.ProjectsPath, host)
		zap.L().Info("WRITE TO HOST", zap.String("Local", setting.Conf.ProjectsPath), zap.String("Host", host))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetGoBuildFilesFromWorker(buildID string) error {
	//for _, host := range setting.Conf.ClusterConfig.Hosts {
	//	_, err := helper.RecvFilesFrom(fmt.Sprintf(setting.Conf.GoBuildPath, buildID), host)
	//	zap.L().Info("GET BUILD FILES FROM WORKER", zap.String("Local", setting.Conf.GoBuildPath), zap.String("Host", host))
	//	if err != nil {
	//		return err
	//	}
	//}
	var wg sync.WaitGroup
	errors := make(chan error, len(setting.Conf.ClusterConfig.Hosts))
	defer close(errors)

	for _, host := range setting.Conf.ClusterConfig.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			_, err := helper.RecvFilesFrom(fmt.Sprintf(setting.Conf.GoBuildPath, buildID), host)
			zap.L().Info("GET BUILD FILES FROM WORKER", zap.String("Local", setting.Conf.GoBuildPath), zap.String("Host", host))
			if err != nil {
				errors <- err
			}
		}(host)
	}
	wg.Wait()
	if len(errors) > 0 {
		return <-errors // Return the first encountered error
	}

	return nil
}

func SendGoBuildFilesToWorker(buildID string) error {
	//for _, host := range setting.Conf.ClusterConfig.Hosts {
	//	_, err := helper.SendFilesTo(fmt.Sprintf(setting.Conf.GoBuildPath, buildID), host)
	//	zap.L().Info("WRITE BUILD FILES TO WORKER", zap.String("Local", setting.Conf.GoBuildPath), zap.String("Host", host))
	//	if err != nil {
	//		return err
	//	}
	//}
	var wg sync.WaitGroup
	errors := make(chan error, len(setting.Conf.ClusterConfig.Hosts))
	defer close(errors)

	for _, host := range setting.Conf.ClusterConfig.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			_, err := helper.SendFilesTo(fmt.Sprintf(setting.Conf.GoBuildPath, buildID), host)
			zap.L().Info("WRITE BUILD FILES TO WORKER", zap.String("Local", setting.Conf.GoBuildPath), zap.String("Host", host))
			if err != nil {
				errors <- err
			}
		}(host)
	}
	wg.Wait()
	if len(errors) > 0 {
		return <-errors // Return the first encountered error
	}
	return nil
}

func HandleCompileGroup(ctx context.Context, t *asynq.Task) error {
	var group parser.CompileGroup
	if err := json.Unmarshal(t.Payload(), &group); err != nil {
		zap.L().Error("HandleCompileJobs json.Unmarshal failed",
			zap.Error(err),
		)
		return err
	}
	var checkList []string
	md5Str := helper.GetMD5Hash(group.ProjectPath + time.Now().String())

	// copy project to cluster
	//if err := SendProjectsToWorker(); err != nil {
	//	return err
	//}
	// copy tmp to cluster
	startTime := time.Now()
	if err := SendGoBuildFilesToWorker(group.BuildPath); err != nil {
		return err
	}
	firstTansferTime := time.Since(startTime).Seconds()

	// start time
	startTime = time.Now()
	// send jobs to cluster
	for _, job := range group.Jobs {
		job.MD5 = md5Str
		checkList = append(checkList, job.Autonomy[0])
		task, err := NewCompileJob(job)
		if err != nil {
			zap.L().Error("failed to create new job ",
				zap.Any("Job", job),
				zap.Error(err),
			)
			return nil
		}
		if _, err = AsynqClient.Enqueue(task, asynq.Queue("Compile_Job"), asynq.Retention(24*time.Hour), asynq.MaxRetry(10)); err != nil {
			zap.L().Error("failed to put task in asynq",
				zap.Any("task", task),
				zap.Error(err),
			)
			return nil
		}
	}
	elapsedTime := time.Since(startTime).Seconds()
	JobDistributedTime += elapsedTime
	fmt.Println("############### Job Distributed Time ###############")
	fmt.Println(JobDistributedTime)
	fmt.Println("####################################################")

	// query redis to get status
	for len(checkList) != 0 {
		checkList = deleteElementFromCheckList(md5Str, checkList)
	}
	// get all files from cluster
	startTime = time.Now()
	if err := GetGoBuildFilesFromWorker(group.BuildPath); err != nil {
		return err
	}
	lastTransferTime := time.Since(startTime).Seconds()
	ExchangeDepTime += lastTransferTime + firstTansferTime
	fmt.Println("############### Dep Exchange Time ###############")
	fmt.Println(ExchangeDepTime)
	fmt.Println("####################################################")
	return nil
}

func deleteElementFromCheckList(md5Str string, values []string) []string {
	var result []string
	for _, v := range values {
		if ok, _ := redis.CheckFileStatue(fmt.Sprintf("%s_%s", md5Str, v)); !ok {
			result = append(result, v)
		}
	}
	return result
}
