package cmd

import (
	"fmt"

	"github.com/PanelMc/worker/infra"
)

func Run() (err error) {
	infra.InitializeLogger()

	var cfg infra.Config
	cfg, err = infra.InitializeConfig()
	if err != nil {
		return
	}
	fmt.Printf("Config: %#v\n", cfg)

	return
}
