package container

import (
	"context"
	"fmt"
	"time"

	"github.com/PanelMc/worker"
)

func (c *dockerContainer) Stop() error {
	c.Logger().Debug("Stopping the server...")
	
	if c.status == worker.StatusStopping {
		return fmt.Errorf("Server already shutting down. Current status: %s", c.status)
	} else if c.status == worker.StatusStopped {
		return fmt.Errorf("Server already stopped. Current status: %s", c.status)
	}

	timeout := time.Duration(time.Second * 15)
	if err := c.client.ContainerStop(context.TODO(), c.ContainerID, &timeout); err != nil {
		c.Logger().Error("Failed to stop the container.")
		return err
	}

	return nil
}