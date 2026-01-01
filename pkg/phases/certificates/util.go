package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func NewPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func NewCACert(caKey *rsa.PrivateKey) ([]byte, error) {
	caCertTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "kubernetes-ca",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, caCertTemplate, caCertTemplate, &caKey.PublicKey, caKey)

	if err != nil {
		return nil, err
	}
	return certBytes, nil
}

type CertOpts struct {
	CommonName   string
	Organization []string
	IPs          []net.IP
	DNSNames     []string
	ExtKeyUsage  []x509.ExtKeyUsage
	IsServerCert bool
}

func NewServerCert(privKey *rsa.PrivateKey, certOpts CertOpts, caCert *x509.Certificate, caKey *rsa.PrivateKey) ([]byte, error) {
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	certTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   certOpts.CommonName,
			Organization: certOpts.Organization,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IPAddresses:           certOpts.IPs,
		DNSNames:              certOpts.DNSNames,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           certOpts.ExtKeyUsage,
		BasicConstraintsValid: true,
	}

	return x509.CreateCertificate(rand.Reader, certTemplate, caCert, &privKey.PublicKey, caKey)
}

func NewClientCert(privKey *rsa.PrivateKey, certOpts CertOpts, caCert *x509.Certificate, caKey *rsa.PrivateKey) ([]byte, error) {
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	certTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   certOpts.CommonName,
			Organization: certOpts.Organization,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
	}

	return x509.CreateCertificate(
		rand.Reader,
		certTemplate,
		caCert,
		&privKey.PublicKey,
		caKey,
	)
}

func createCertificate(component string, certOpts CertOpts) error {
	privKey, err := NewPrivateKey()
	if err != nil {
		return fmt.Errorf("%s key generation failed: %s", component, err)
	}

	caCert, caKey, err := loadCA()
	if err != nil {
		return fmt.Errorf("%s key generation failed: %s", component, err)
	}

	var certBytes []byte

	if certOpts.IsServerCert {
		certBytes, err = NewServerCert(privKey, certOpts, caCert, caKey)
	} else {
		certBytes, err = NewClientCert(privKey, certOpts, caCert, caKey)
	}
	if err != nil {
		return fmt.Errorf("%s cert generation failed: %s", component, err)
	}

	keyOut, err := os.Create(fmt.Sprintf("/etc/kubernetes/pki/%s.key", component))
	if err != nil {
		return fmt.Errorf("%s key saving failed: %s", component, err)
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})
	os.Chmod(fmt.Sprintf("/etc/kubernetes/pki/%s.key", component), 0600)

	crtOut, err := os.Create(fmt.Sprintf("/etc/kubernetes/pki/%s.crt", component))
	if err != nil {
		return fmt.Errorf("%s cert saving failed: %s", component, err)
	}
	defer crtOut.Close()
	pem.Encode(crtOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	os.Chmod(fmt.Sprintf("/etc/kubernetes/pki/%s.crt", component), 0600)

	fmt.Printf("[certificate] %s certificate successfully generated\n", component)

	return nil
}

func createServiceAccountKeys() error {
	saKey, err := NewPrivateKey()
	if err != nil {
		return fmt.Errorf("SA key generation failed: %s", err)
	}

	keyOut, err := os.Create("/etc/kubernetes/pki/sa.key")
	if err != nil {
		return fmt.Errorf("SA key saving failed: %s", err)
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(saKey),
	})
	os.Chmod("/etc/kubernetes/pki/sa.key", 0600)

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&saKey.PublicKey)
	if err != nil {
		return fmt.Errorf("SA public key marshaling failed: %s", err)
	}

	pubOut, err := os.Create("/etc/kubernetes/pki/sa.pub")
	if err != nil {
		return fmt.Errorf("SA public key saving failed: %s", err)
	}
	defer pubOut.Close()
	pem.Encode(pubOut, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	fmt.Println("[certificate] Service account keys successfully generated")
	return nil
}
