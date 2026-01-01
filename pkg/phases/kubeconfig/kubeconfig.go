package kubeconfig

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func SetupKubeconfigs(advertiseAddress string) error {
	err := CreateKubeconfig(
		"/etc/kubernetes/admin.conf",
		"kubernetes",
		"kubernetes-admin",
		"/etc/kubernetes/pki/admin.crt",
		"/etc/kubernetes/pki/admin.key",
		advertiseAddress,
	)
	if err != nil {
		return err
	}

	err = CreateKubeconfig(
		"/etc/kubernetes/scheduler.conf",
		"kubernetes",
		"system:kube-scheduler",
		"/etc/kubernetes/pki/scheduler.key",
		"/etc/kubernetes/pki/scheduler.crt",
		advertiseAddress,
	)
	if err != nil {
		return err
	}

	err = CreateKubeconfig(
		"/etc/kubernetes/controller-manager.conf",
		"kubernetes",
		"system:kube-controller-manager",
		"/etc/kubernetes/pki/controller-manager.crt",
		"/etc/kubernetes/pki/controller-manager.key",
		advertiseAddress,
	)
	if err != nil {
		return err
	}

	return nil
}

func CreateKubeconfig(
	kubeconfigPath string,
	clusterName string,
	user string,
	certPath string,
	keyPath string,
	advertiseAddress string,
) error {

	kubeconfig := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				Server:               fmt.Sprintf("https://%s:6443", advertiseAddress),
				CertificateAuthority: "/etc/kubernetes/pki/ca.crt",
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			user: {
				ClientKey:         keyPath,
				ClientCertificate: certPath,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			clusterName: {
				Cluster:  clusterName,
				AuthInfo: user,
			},
		},
		CurrentContext: fmt.Sprintf("%s@%s", user, clusterName),
	}

	err := clientcmd.WriteToFile(kubeconfig, kubeconfigPath)
	if err != nil {
		return err
	}

	fmt.Printf("[kubeconfig] %s kubeconfig successfully generated\n", user)
	return nil
}
