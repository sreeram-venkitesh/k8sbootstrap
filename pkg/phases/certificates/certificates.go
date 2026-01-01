package certificates

import (
	"crypto/x509"
	"fmt"
	"net"
	"os"
)

func SetupCerts(advertiseAddress string) error {
	err := createKubernetesCA()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	kubeApiserverIPs := []net.IP{
		net.IPv4(127, 0, 0, 1),
		net.IPv4(10, 96, 0, 1), //default pod CIDR IP of kubeadm
		net.ParseIP(advertiseAddress),
	}
	kubeApiserverDNS := []string{
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster.local",
		hostname,
	}
	err = createCertificate(
		"kube-apiserver",
		CertOpts{
			CommonName: "kube-apiserver",
			IPs:        kubeApiserverIPs,
			DNSNames:   kubeApiserverDNS,
			ExtKeyUsage: []x509.ExtKeyUsage{
				x509.ExtKeyUsageServerAuth,
			},
			IsServerCert: true,
		},
	)
	if err != nil {
		return err
	}

	err = createCertificate("apiserver-kubelet-client", CertOpts{
		CommonName:   "apiserver-kubelet-client",
		Organization: []string{"system:masters"},
	})
	if err != nil {
		return err
	}

	err = createCertificate("apiserver-etcd-client", CertOpts{
		CommonName:   "kube-apiserver-etcd-client",
		Organization: []string{"system:masters"},
	})
	if err != nil {
		return err
	}

	err = createCertificate("controller-manager", CertOpts{
		CommonName: "system:kube-controller-manager",
	})
	if err != nil {
		return err
	}

	err = createCertificate("scheduler", CertOpts{
		CommonName: "system:kube-scheduler",
	})
	if err != nil {
		return err
	}

	err = createCertificate("admin", CertOpts{
		CommonName:   "kubernetes-admin",
		Organization: []string{"system:masters"},
	})
	if err != nil {
		return err
	}

	err = createCertificate("etcd-server", CertOpts{
		CommonName: "etcd-server",
		IPs: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP(advertiseAddress),
		},
		DNSNames: []string{
			"localhost",
			hostname,
		},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		IsServerCert: true,
	})
	if err != nil {
		return err
	}

	err = createServiceAccountKeys()
	if err != nil {
		return err
	}

	return nil
}
