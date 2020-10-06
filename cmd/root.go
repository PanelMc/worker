package cmd

import (
	"errors"
	"fmt"

	"github.com/PanelMc/worker"
	"github.com/PanelMc/worker/container"
	"github.com/PanelMc/worker/infra"
	"github.com/sirupsen/logrus"
)

func Run() error {
	infra.InitializeLogger()

	c, err := container.NewDockerContainer("Teste", worker.ContainerOptions{
		Image: "itzg/minecraft-server",
		Ports: []int{25565},
		RAM:   "1GB",
		Swap:  "1GB",
	})
	if err != nil {
		logrus.WithError(err).Errorln("Failed to create a new container")
		return err
	}

	s, err := c.Stats()
	if err != nil {
		logrus.WithError(err).Errorln("Failed to create a new container")
	} else {
		logrus.Infof("Stats: %v", s)
	}

	if err := c.Start(); err != nil {
		logrus.WithError(err).Errorln("Failed to start the new container")
	}

	return nil
}
