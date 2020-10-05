package container

import (
	"context"
	"fmt"

	"github.com/PanelMc/worker"
	"github.com/docker/docker/api/types"
)

func (c *dockerContainer) Start() error {
	c.logger.Debug("Starting the container...")

	if c.status != worker.StatusStopped {
		return fmt.Errorf("Server already running. Current status: %s", c.status)
	}

	if err := c.client.ContainerStart(context.TODO(), c.ContainerID, types.ContainerStartOptions{}); err != nil {
		c.Logger().Error("Failed to start the container.")
		return err
	}

	c.status = worker.StatusRunning
	c.Logger().Info("Container started.")
	return nil
}
