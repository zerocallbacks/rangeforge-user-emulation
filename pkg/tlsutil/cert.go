package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// EnsureCertificates verifies certificate and key files exist, or automatically generates them.
func EnsureCertificates(certPath, keyPath string) (tls.Certificate, error) {
	if certPath == "" {
		certPath = filepath.Join("certs", "server.crt")
	}
	if keyPath == "" {
		keyPath = filepath.Join("certs", "server.key")
	}

	// Check if already exist and non-empty
	certInfo, certErr := os.Stat(certPath)
	keyInfo, keyErr := os.Stat(keyPath)

	needsGen := os.IsNotExist(certErr) || os.IsNotExist(keyErr) ||
		(certInfo != nil && certInfo.Size() == 0) ||
		(keyInfo != nil && keyInfo.Size() == 0)

	if needsGen {
		if err := GenerateSelfSignedCert(certPath, keyPath); err != nil {
			return tls.Certificate{}, fmt.Errorf("generating TLS certificates: %w", err)
		}
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		// Existing files were corrupted or invalid; regenerate as self-healing fallback
		if genErr := GenerateSelfSignedCert(certPath, keyPath); genErr != nil {
			return tls.Certificate{}, fmt.Errorf("regenerating corrupted TLS certificates: %w (original error: %v)", genErr, err)
		}
		return tls.LoadX509KeyPair(certPath, keyPath)
	}

	return cert, nil
}

// GenerateSelfSignedCert creates a high-security ECDSA P-256 certificate for cyber range,
// lab, enterprise, and multi-network use.
func GenerateSelfSignedCert(certPath, keyPath string) error {
	// Create directory if missing
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0755); err != nil {
		return err
	}

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generating ecdsa key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("generating serial number: %w", err)
	}

	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := notBefore.Add(365 * 24 * time.Hour) // 1 year validity

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cyber Range Operations"},
			CommonName:   "RangeForge Gateway",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
			net.ParseIP("10.0.0.1"),
			net.ParseIP("192.168.1.1"),
			net.ParseIP("192.168.0.1"),
			net.ParseIP("172.16.0.1"),
		},
		DNSNames: []string{
			"localhost",
			"*.range.local",
			"*.corp.local",
			"*.local",
			"*.lan",
			"*.internal",
			"*.home.arpa",
			"gateway.range.local",
		},
	}

	if h, err := os.Hostname(); err == nil && h != "" {
		template.DNSNames = append(template.DNSNames, h)
	}

	// Also append all detected network interface IPs
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP != nil {
					template.IPAddresses = append(template.IPAddresses, ipNet.IP)
				}
			}
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return fmt.Errorf("creating x509 certificate: %w", err)
	}

	// Write Certificate PEM
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("creating cert file %s: %w", certPath, err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("writing cert pem: %w", err)
	}

	// Write Private Key PEM
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("creating key file %s: %w", keyPath, err)
	}
	defer keyOut.Close()

	keyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("marshaling ecdsa private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return fmt.Errorf("writing key pem: %w", err)
	}

	return nil
}
