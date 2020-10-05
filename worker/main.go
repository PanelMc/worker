package main

import (
	"fmt"
	"os"

	"github.com/PanelMc/worker/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error ocurred during execution: %s\n", err.Error())
		os.Exit(1)
	}
}
