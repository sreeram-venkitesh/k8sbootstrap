package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sreeram-venkitesh/k8sbootstrap/pkg/phases/preflight"
)

func newCmdInit() *cobra.Command {
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Run this command in order to set up the Kubernetes control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[init] Starting k8sbootstrap init...")

			if err := preflight.RunPreflightChecks(); err != nil {
				fmt.Printf("[preflight] Preflight checks failed: %s\n", err)
			}

			return nil
		},
	}

	return initCmd
}
