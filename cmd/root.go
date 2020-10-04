package cmd

import "github.com/PanelMc/worker/infra"

func Run() error {
	infra.InitializeLogger()

	return nil
}
