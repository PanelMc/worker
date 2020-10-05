package worker

import (
	"github.com/sirupsen/logrus"
)

// Container represents the container the server is running on.
type Container interface {
	// Start starts the container if not running already
	Start() error
	// Stop stopps the container if running
	Stop() error
	// Exec executes a command on the container
	Exec(cmd string) error
	// Stats returns the last stats obtained from the container
	Stats() (ContainerStats, error)
	// StatsChan returns a channel that receives the container stats
	StatsChan() (<-chan *ContainerStats, error)

	Logger() *logrus.Logger
}

type ContainerStats struct {
	CPUPercentage    float64 `json:"cpu_percentage"`
	MemoryPercentage float64 `json:"memory_percentage"`
	Memory           uint64  `json:"memory"`
	MemoryLimit      uint64  `json:"memory_limit"`
	NetworkDownload  uint64  `json:"network_download"`
	NetworkUpload    uint64  `json:"network_upload"`
	DiscRead         uint64  `json:"disc_read"`
	DiscWrite        uint64  `json:"disc_write"`
}

type ContainerOptions struct {
	Image string `json:"image,omitempty"`
	Ram   string `json:"ram,omitempty"`
	Swap  string `json:"swap,omitempty"`
	Ports []int  `json:"ports,omitempty"`
}
