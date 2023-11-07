package main

import (
	"distgo/internal/router"
	"distgo/internal/setting"
	"fmt"
)

const ConfigFilePath = "configs/web.yaml"

func main() {

	if err := setting.Init(ConfigFilePath); err != nil {
		fmt.Printf("load config failed, err:%v\n", err)
	}

	err, r := router.SetupRouter(setting.Conf.Mode)
	if err != nil {
		fmt.Printf("setup router failed, err:%v\n", err)
	}

	err = r.Run(fmt.Sprintf(":%d", setting.Conf.Port))
	if err != nil {
		fmt.Printf("run web service failed, err:%v\n", err)
	}
}
