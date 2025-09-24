// entry point to app :)
package main

import (
	"github.com/ds124wfegd/WB_L2/18/config"
	"github.com/ds124wfegd/WB_L2/18/internal/appServer"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	viperInstance, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Cannot load config. Error: {%s}", err.Error())
	}

	cfg, err := config.ParseConfig(viperInstance)
	if err != nil {
		logrus.Fatalf("Cannot parse config. Error: {%s}", err.Error())
	}

	appServer.NewServer(cfg)
}
