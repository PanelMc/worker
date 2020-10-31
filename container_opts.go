package worker

import (
	"strings"
)

// ContainerOpts helps you to easily model the container options
// according to your needs.
type ContainerOpts func(*ContainerOptions)

func WithPreset(preset ServerPreset) ContainerOpts {
	return func(co *ContainerOptions) {
		cID := strings.TrimSpace(preset.ServerID)
		if cID != "" {
			co.ContainerName = cID
		}

		if preset.ContainerImage != nil {
			co.Image = *preset.ContainerImage
		}

		if preset.Memory != nil {
			if preset.Memory.Limit != "" {
				co.Memory.Limit = preset.Memory.Limit
			}

			if preset.Memory.Swap != "" {
				co.Memory.Swap = preset.Memory.Swap
			}
		}

		if preset.Network != nil {
			co.Network = preset.Network
		}

		if len(preset.Binds) > 0 {
			co.Binds = preset.Binds
		}
	}
}
