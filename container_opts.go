package worker

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
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
			cImage := strings.TrimSpace(preset.ContainerImage.ID)
			if cImage != "" {
				co.Image = cImage
			}
		}

		if preset.Memory != nil {
			cMem := strings.TrimSpace(preset.Memory.Limit)
			if cMem != "" {
				co.RAM = cMem
			}

			cSwap := strings.TrimSpace(preset.Memory.Swap)
			if cSwap != "" {
				co.Swap = cSwap
			}
		}

		if preset.Network != nil {
			cExpose := preset.Network.Expose
			exposeLen := len(cExpose)
			if exposeLen > 0 {
				for _, e := range cExpose {
					port := e[strings.IndexRune(e, ':'):]
					p, err := strconv.Atoi(port)
					if err != nil {
						logrus.Warnf("Failed to parse the port %s to integer! (source: %s)", port, e)
						continue
					}

					co.Ports = append(co.Ports, p)
				}
			}
		}
	}
}
