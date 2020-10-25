package infra

import (
	"os"

	"github.com/PanelMc/worker/io"
	"github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
)

// InitializeLogger initializes the global logger with the default configuration
func InitializeLogger() {
	logrus.SetFormatter(&nested.Formatter{
		HideKeys: true,
		NoColors: true,
	})
	logrus.SetLevel(logrus.TraceLevel)
}

func InitializeConfig() (cfg Config, err error) {
	var c config

	err = io.LoadConfig("config.hcl", &c)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}

		err = io.SaveConfig(config{
			Server: &struct {
				Binds []struct {
					HostDir string "hcl:\"host_dir\""
					Volume  string "hcl:\"volume\""
				} "hcl:\"bind,block\""
			}{
				Binds: []struct {
					HostDir string "hcl:\"host_dir\""
					Volume  string "hcl:\"volume\""
				}{
					{
						HostDir: "/etc/worker/%s/data/",
						Volume:  "/data",
					},
				},
			},
			PresetsFolder:     "./presets/",
			FilePermissions:   644,
			FolderPermissions: 744,
		}, "config.hcl")
		if err != nil {
			return
		}

		return InitializeConfig()
	}

	var serverConfig *ServerConfig
	if c.Server != nil {
		serverConfig = &ServerConfig{
			Binds: make([]ServerBindConfig, len(c.Server.Binds)),
		}
	}

	cfg = Config{
		Server: serverConfig,

		PresetsFolder:     c.PresetsFolder,
		FilePermissions:   c.FilePermissions,
		FolderPermissions: c.FolderPermissions,
	}

	if serverConfig != nil {
		// Map the serverConfig
		for i, bind := range c.Server.Binds {
			cfg.Server.Binds[i] = ServerBindConfig{
				HostDir: bind.HostDir,
				Volume:  bind.Volume,
			}
		}
	}

	return
}

type config struct {
	// Server as a struct array, making it optional
	Server *struct {
		Binds []struct {
			HostDir string `hcl:"host_dir"`
			Volume  string `hcl:"volume"`
		} `hcl:"bind,block"`
	} `hcl:"server,block"`

	PresetsFolder     string      `hcl:"presets_folder"`
	FilePermissions   os.FileMode `hcl:"file_permissions"`
	FolderPermissions os.FileMode `hcl:"folder_permissions"`
}

// Config defines how the worker should run
type Config struct {
	// Servere defines default configuration for new servers created
	Server *ServerConfig

	// PresetsFolder defines the folder to be used for server preset files.
	PresetsFolder string
	// Permission used when creating a new file. e.g. configuration files
	FilePermissions os.FileMode
	// Permission used when creating a new folder
	FolderPermissions os.FileMode
}

// ServerConfig defines default configuration for new servers created
type ServerConfig struct {
	// Binds defines wich volume binds to use.
	Binds []ServerBindConfig
}

// ServerBindConfig defines wich volume binds to use.
type ServerBindConfig struct {
	// HostDir defines where to bind the volume on the host machine.
	HostDir string
	// Volume defines the volume to be binded.
	Volume string
}
