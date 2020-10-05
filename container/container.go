package container

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"code.cloudfoundry.org/bytefmt"
	"github.com/PanelMc/worker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

type dockerContainer struct {
	sync.Mutex

	ServerID    string
	ContainerID string
	status      worker.Status

	client    *client.Client
	statsChan <-chan *worker.ContainerStats

	logger *logrus.Logger
}

func (c *dockerContainer) Logger() *logrus.Logger {
	return c.logger
}

// Ensure the container struct implements the Container
// interface. If not, the program won't compile.
var _ worker.Container = &dockerContainer{}

var logger = logrus.WithField("context", "container").Logger

// NewDockerContainer creates a new docker container using the given options
func NewDockerContainer(serverID string, opts worker.ContainerOptions) (worker.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	logger := logrus.WithField("container", serverID).Logger
	container := &dockerContainer{
		ServerID: serverID,
		status:   worker.StatusStopped,
		client:   cli,
		logger:   logger,
	}

	ctx := context.TODO()

	if err := prepare(ctx, container, &opts); err != nil {
		return nil, err
	}

	containerConfig := parseContainerConfig(serverID, &opts)
	containerHostConfig := parseHostConfig(serverID, &opts)

	resContainer, err := cli.ContainerCreate(ctx, &containerConfig, &containerHostConfig, nil, containerConfig.Hostname)
	if err != nil {
		return nil, err
	}

	container.ContainerID = resContainer.ID

	return container, nil
}

func prepare(ctx context.Context, container *dockerContainer, opts *worker.ContainerOptions) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	progress, err := pullImage(ctx, container, opts.Image)
	if err != nil {
		return err
	}

	var p *imagePullEvent
	var ok bool
	for ok {
		select {
		case <-ctx.Done():
			return nil
		case p, ok = <-progress:
			if ok {
				container.Logger().Debugf("Pulling progress: %d/%d - %s", p.ProgressDetail.Current, p.ProgressDetail.Total, p.Status)
			}
		}
	}

	return nil
}

func parseContainerConfig(serverID string, opts *worker.ContainerOptions) container.Config {
	portSet := nat.PortSet{}
	for _, p := range opts.Ports {
		port := nat.Port(fmt.Sprintf("%d/%s", p, "tcp"))
		portSet[port] = struct{}{}
	}

	containerConfig := container.Config{
		Image:        opts.Image,
		AttachStdin:  true,
		OpenStdin:    true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Hostname:     "daemon-" + serverID,
		ExposedPorts: portSet,
		Volumes: map[string]struct{}{
			"/data": {},
		},
		Env: []string{
			"EULA=TRUE",
			"PAPER_DOWNLOAD_URL=https://heroslender.com/assets/PaperSpigot-1.8.8.jar",
			"TYPE=PAPER",
			"VERSION=1.8.8",
			"ENABLE_RCON=false",
		},
	}

	return containerConfig
}

func parseHostConfig(serverID string, opts *worker.ContainerOptions) container.HostConfig {
	portMap := nat.PortMap{}
	for _, p := range opts.Ports {
		port := nat.Port(fmt.Sprintf("%d/%s", p, "tcp"))
		portMap[port] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", p)}}
	}

	// fix windows path
	// path := strings.Replace(c.server.DataPath(), "C:\\", "/c/", 1)
	path := "/home/heroslender/projects/PanelMc/worker/" + serverID
	path = strings.Replace(path, "\\", "/", -1)
	// point to `/data` volume
	path += ":/data"

	memory, err := bytefmt.ToBytes(opts.Ram)
	if err != nil {
		logrus.Error("Failed to read server RAM, using default(1 Gigabyte).")
		memory = 1073741824 // 1GB Default
	}
	swap, err := bytefmt.ToBytes(opts.Swap)
	if err != nil {
		logrus.Error("Failed to read server Swap, using default(1 Gigabyte).")
		swap = 1073741824 // 1GB Default
	}

	containerHostConfig := container.HostConfig{
		Resources: container.Resources{
			Memory:     int64(memory),
			MemorySwap: int64(swap),
		},
		Binds:        []string{path},
		PortBindings: portMap,
	}

	return containerHostConfig
}
