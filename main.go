package main

import (
	"os"

	"github.com/sreeram-venkitesh/k8sbootstrap/cmd"
)

func main() {
	cmd := cmd.NewK8sBootstrapCmd()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
