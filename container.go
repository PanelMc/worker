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
	// Status says whether the server is running or not
	Status() Status
	// Logger returns the logger used by the server
	// logs sent here, will be redirected to the container stdout
	Logger() *logrus.Entry
}

// ContainerStats holds the stats relative to a container
// at a point in time.
type ContainerStats struct {
	// Percentage of CPU usage, sum of all cores
	CPUPercentage float64 `json:"cpu_percentage"`
	// Percentage of RAM usage
	MemoryPercentage float64 `json:"memory_percentage"`
	// RAM usage in bytes
	Memory uint64 `json:"memory"`
	// Max available RAM in bytes
	MemoryLimit uint64 `json:"memory_limit"`
	// Total network download bytes, since start
	NetworkDownload uint64 `json:"network_download"`
	// Total network upload bytes, since start
	NetworkUpload uint64 `json:"network_upload"`
	// Disc read
	// TODO investigate meaning of returned values
	DiscRead uint64 `json:"disc_read"`
	// Disc write
	DiscWrite uint64 `json:"disc_write"`
}

// ContainerOptions holds the options used to create a new container
type ContainerOptions struct {
	ContainerName  string `json:"container_name,omitempty"`
	Binds          []ContainerBind
	Image ContainerImage    `json:"container_image"`
	Memory         ContainerMemory   `json:"memory"`
	Network        *ContainerNetwork `json:"network,omitempty"`
}

type ContainerImage struct {
	ID string `hcl:"id" json:"id"`
}

type ContainerMemory struct {
	Limit string `hcl:"limit" json:"limit"`
	Swap  string `hcl:"swap,optional" json:"swap"`
}

type ContainerNetwork struct {
	Expose []string `hcl:"expose,optional" json:"expose"`
}

// ContainerBind defines which volume binds to use.
type ContainerBind struct {
	// HostDir defines where to bind the volume on the host machine.
	HostDir string `hcl:"host_dir" json:"host_dir,omitempty"`
	// Volume defines the volume to be binded.
	Volume string `hcl:"volume,optional" json:"volume"`
}
