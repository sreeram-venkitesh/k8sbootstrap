package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sreeram-venkitesh/k8sbootstrap/pkg/phases/certificates"
	"github.com/sreeram-venkitesh/k8sbootstrap/pkg/phases/preflight"
)

var (
	advertiseAddress string
	podNetworkCIDR   string
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

			if err := certificates.SetupCerts(advertiseAddress); err != nil {
				fmt.Printf("[certificate] Certificate creation failed: %s\n", err)
			}

			return nil
		},
	}

	initCmd.Flags().StringVar(
		&advertiseAddress,
		"advertise-address",
		"",
		"The IP address the API Server will advertise it's listening on",
	)
	initCmd.Flags().StringVar(
		&podNetworkCIDR,
		"pod-network-cidr",
		"10.244.0.0/16",
		"Specify range of IP addresses for the pod network",
	)

	initCmd.MarkFlagRequired("advertise-address")

	return initCmd
}
