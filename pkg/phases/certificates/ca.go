package certificates

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

func createKubernetesCA() error {
	caKey, err := NewPrivateKey()
	if err != nil {
		return fmt.Errorf("CA key generation failed: %s", err)
	}

	caCertBytes, err := NewCACert(caKey)
	if err != nil {
		return fmt.Errorf("CA cert generation failed: %s", err)
	}

	err = os.MkdirAll("/etc/kubernetes/pki", 0755)
	if err != nil {
		return fmt.Errorf("Failed to create /etc/kubernetes: %s", err)
	}

	keyOut, err := os.Create("/etc/kubernetes/pki/ca.key")
	if err != nil {
		return fmt.Errorf("CA key saving failed: %s", err)
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	})
	os.Chmod("/etc/kubernetes/pki/ca.key", 0600)

	crtOut, err := os.Create("/etc/kubernetes/pki/ca.crt")
	if err != nil {
		return fmt.Errorf("CA cert saving failed: %s", err)
	}
	defer crtOut.Close()
	pem.Encode(crtOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	})
	os.Chmod("/etc/kubernetes/pki/ca.crt", 0600)

	fmt.Println("[certificate] CA certificate successfully generated")

	return nil
}

func loadCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	caCertPEM, err := ioutil.ReadFile("/etc/kubernetes/pki/ca.crt")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load CA cert: %w", err)
	}

	block, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA cert: %w", err)
	}

	caKeyPEM, err := ioutil.ReadFile("/etc/kubernetes/pki/ca.key")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load CA key: %w", err)
	}

	block, _ = pem.Decode(caKeyPEM)
	caKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA key: %w", err)
	}

	return caCert, caKey, nil
}
