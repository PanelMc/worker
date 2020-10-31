package worker

// Server represents a game server.
type Server interface {
	Start() error

	Stop() error

	SendCommand(cmd string) error
}

type server struct {
	container Container
}

// Status represents whether the server is running or not.
type Status string

const (
	// StatusRunning indicates the server is running.
	StatusRunning Status = "running"
	// StatusStopped indicates the server is stopped.
	StatusStopped Status = "stopped"
	// StatusStarting indicates the server is still starting and not yet running.
	// This option can be omitted and passed directly to running if
	// the server software in use is not supported.
	StatusStarting Status = "starting"
	// StatusStopping indicates the server is stopping, but not yet stopped.
	StatusStopping Status = "stopping"
)

// NewServer initializes a new Server instance based on the provided Container.
func NewServer(container Container) (Server, error) {
	return &server{container}, nil
}

// ServerCreateOptions holds information needed to create a new Server
type ServerCreateOptions struct {
	ServerID   string `hcl:"server_id"`
	ServerName string `hcl:"server_name"`

	// Binds defines which volume binds to use.
	Binds          []ContainerBind
	ContainerImage *ContainerImage `hcl:"container_image,block"`
	Memory         *ContainerMemory         `hcl:"memory,block"`
	Network        *ContainerNetwork        `hcl:"network,block"`
}

// ServerPreset represents a preset to be used for Server creation
type ServerPreset ServerCreateOptions
