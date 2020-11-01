package container

import (
	"context"
	"errors"
	"fmt"
	"net"
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

	// ContainerName is a unique, user readable, identifier
	// of the container.
	ContainerName string
	ContainerID   string
	status        worker.Status

	client    *client.Client
	statsChan <-chan *worker.ContainerStats

	logger *logrus.Entry
}

func (c *dockerContainer) Logger() *logrus.Entry {
	return c.logger
}

func (c *dockerContainer) Status() worker.Status {
	return c.status
}

var logger = logrus.WithField("context", "container")

// NewDockerContainer creates a new docker container using the given options
func NewDockerContainer(opts ...worker.ContainerOpts) (worker.Container, error) {
	options := &worker.ContainerOptions{
		ContainerName: "minecraft",
		Image: worker.ContainerImage{
			ID: "itzg/minecraft-server",
		},
		Memory: worker.ContainerMemory{
			Limit: "1GB",
			Swap:  "1GB",
		},
		Binds: make([]worker.ContainerBind, 0),
		Network: &worker.ContainerNetwork{
			Binds: make([]worker.ContainerNetworkBind, 0),
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	logger := logrus.WithField("container", options.ContainerName)
	container := &dockerContainer{
		ContainerName: options.ContainerName,
		status:        worker.StatusStopped,
		client:        cli,
		logger:        logger,
	}

	ctx := context.TODO()

	if err := prepare(ctx, container, options); err != nil {
		return nil, err
	}

	containerConfig := parseContainerConfig(container, options)
	containerHostConfig := parseHostConfig(container, options)

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

	container.Logger().Debugln("Checking image for updates...")
	progress, err := pullImage(ctx, container, opts.Image)
	if err != nil {
		return fmt.Errorf("image pull error for '%s': %w", opts.Image, err)
	}

	var p *imagePullEvent
	var ok bool = true
	for ok {
		select {
		case <-ctx.Done():
			return nil
		case p, ok = <-progress:
			if ok {
				if p.Progress != "" {
					container.Logger().Debugf("Pulling progress: %s | %s", p.Status, p.Progress)
				} else {
					container.Logger().Debugf("Pulling progress: %s", p.Status)
				}
			}
		}
	}

	// Return error from last message if present
	if p != nil && p.Error != "" {
		return errors.New(p.Error)
	}

	return nil
}

func parseContainerConfig(c *dockerContainer, opts *worker.ContainerOptions) container.Config {
	portSet, _, err := parsePortSpecs(opts.Network.Binds)
	if err != nil {
		c.Logger().Errorf("Error parsing the container ports: %s", err)
		portSet = make(map[nat.Port]struct{})
	}

	volumes, _ := parseVolumeBinds(c, opts.Binds)

	containerConfig := container.Config{
		Image:        opts.Image.ID,
		AttachStdin:  true,
		OpenStdin:    true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Hostname:     "daemon-" + c.ContainerName,
		ExposedPorts: portSet,
		Volumes:      volumes,
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

func parseHostConfig(c *dockerContainer, opts *worker.ContainerOptions) container.HostConfig {
	_, portMap, err := parsePortSpecs(opts.Network.Binds)
	if err != nil {
		c.Logger().Errorf("Error parsing the container ports: %s", err)
		portMap = make(map[nat.Port][]nat.PortBinding)
	}

	_, binds := parseVolumeBinds(c, opts.Binds)

	memory, err := bytefmt.ToBytes(opts.Memory.Limit)
	if err != nil {
		c.Logger().Error("Failed to read server RAM, using default(1 Gigabyte).")
		memory = 1073741824 // 1GB Default
	}
	swap, err := bytefmt.ToBytes(opts.Memory.Swap)
	if err != nil {
		c.Logger().Error("Failed to read server Swap, using default(1 Gigabyte).")
		swap = 1073741824 // 1GB Default
	}

	containerHostConfig := container.HostConfig{
		Resources: container.Resources{
			Memory:     int64(memory),
			MemorySwap: int64(swap),
		},
		Binds:        binds,
		PortBindings: portMap,
	}

	return containerHostConfig
}

func parsePortSpecs(portBinds []worker.ContainerNetworkBind) (nat.PortSet, nat.PortMap, error) {
	var (
		exposedPorts = make(nat.PortSet)
		bindings     = make(nat.PortMap)
	)

	for _, rawPort := range portBinds {
		portMappings, err := parsePortSpec(rawPort)
		if err != nil {
			return nil, nil, err
		}

		for _, portMapping := range portMappings {
			port := portMapping.Port
			if _, exists := exposedPorts[port]; !exists {
				exposedPorts[port] = struct{}{}
			}
			bSlice, exists := bindings[port]
			if !exists {
				bSlice = []nat.PortBinding{}
			}
			bindings[port] = append(bSlice, portMapping.Binding)
		}
	}

	logrus.Infof("Ports: %#v", bindings)
	return exposedPorts, bindings, nil
}

func parsePortSpec(portBind worker.ContainerNetworkBind) ([]nat.PortMapping, error) {
	hostIP, hostPort := splitIPPort(portBind.Addr)
	hostIP, err := parseIP(hostIP)
	if err != nil {
		return nil, err
	}

	if hostPort == "" {
		return nil, errors.New("undefined port for bind")
	}

	cPort := portBind.Private
	if cPort == "" {
		cPort = hostPort
	}

	cProto := portBind.Proto
	if cProto == "" {
		cProto = "tcp"
	}

	port, err := nat.NewPort(cProto, cPort)
	if err != nil {
		return nil, err
	}

	// TODO port ranges
	return []nat.PortMapping{
		{
			Port: port,
			Binding: nat.PortBinding{
				HostIP:   hostIP,
				HostPort: hostPort,
			},
		},
	}, nil
}

func parseIP(rawIP string) (string, error) {
	// Strip [] from IPV6 addresses
	ip, _, err := net.SplitHostPort(rawIP + ":")
	if err != nil {
		return "", fmt.Errorf("invalid ip address %v: %s", rawIP, err)
	}
	if ip != "" && net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid ip address: %s", ip)
	}

	return ip, nil
}

func splitIPPort(rawIP string) (string, string) {
	parts := strings.Split(rawIP, ":")
	length := len(parts)

	switch length {
	case 1:
		return "", parts[0]
	case 2:
		return parts[0], parts[1]
	default:
		// IPV6
		return strings.Join(parts[:length-1], ":"), parts[length-1]
	}
}

func parseVolumeBinds(c *dockerContainer, binds []worker.ContainerBind) (map[string]struct{}, []string) {
	var (
		volumes  = make(map[string]struct{})
		bindings = make([]string, 0)
	)

	for _, bind := range binds {
		bind.HostDir = strings.ReplaceAll(bind.HostDir, "{id}", c.ContainerName)
		volumes[bind.Volume] = struct{}{}
		binding := fmt.Sprintf("%s:%s", bind.HostDir, bind.Volume)

		var contains bool
		for _, b := range bindings {
			if b == binding {
				contains = true
				break
			}
		}

		if !contains {
			bindings = append(bindings, binding)
		}
	}

	return volumes, bindings
}
