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
	"time"
)

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

func GetGoBuildFilesFromWorker() error {
	for _, host := range setting.Conf.ClusterConfig.Hosts {
		_, err := helper.RecvFilesFrom(setting.Conf.GoBuildPath, host)
		zap.L().Info("GET BUILD FILES FROM WORKER", zap.String("Local", setting.Conf.GoBuildPath), zap.String("Host", host))
		if err != nil {
			return err
		}
	}
	return nil
}

func SendGoBuildFilesToWorker() error {
	for _, host := range setting.Conf.ClusterConfig.Hosts {
		_, err := helper.SendFilesTo(setting.Conf.GoBuildPath, host)
		zap.L().Info("WRITE BUILD FILES TO WORKER", zap.String("Local", setting.Conf.GoBuildPath), zap.String("Host", host))
		if err != nil {
			return err
		}
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
	if err := SendProjectsToWorker(); err != nil {
		return err
	}
	// copy tmp to cluster
	if err := SendGoBuildFilesToWorker(); err != nil {
		return err
	}

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
	// query redis to get status
	for len(checkList) != 0 {
		checkList = deleteElementFromCheckList(md5Str, checkList)
	}
	// get all files from cluster
	if err := GetGoBuildFilesFromWorker(); err != nil {
		return err
	}

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
