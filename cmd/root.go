package cmd

import (
	"github.com/spf13/cobra"
)

func NewK8sBootstrapCmd() *cobra.Command {
	var cmds = &cobra.Command{
		Use:   "k8sbootstrap",
		Short: "k8sbootstrap is a toy implementation of kubeadm",
		Long:  "A minimal Kubernetes cluster bootstrap tool inspired by kubeadm",
	}

	cmds.AddCommand(newCmdInit())
	return cmds
}
