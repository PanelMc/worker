package worker

// Server represents a game server.
type Server interface {
	Start() error

	Stop() error

	SendCommand(cmd string) error

	Ping() (int, error)
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
	// StatusStopping indicates the server is stopping, but not yet stopeed.
	StatusStopping Status = "stopping"
)
