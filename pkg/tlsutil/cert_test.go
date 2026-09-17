package tlsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateAndEnsureCertificates(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "certs", "test_server.crt")
	keyPath := filepath.Join(tmpDir, "certs", "test_server.key")

	cert, err := EnsureCertificates(certPath, keyPath)
	if err != nil {
		t.Fatalf("EnsureCertificates failed: %v", err)
	}

	if len(cert.Certificate) == 0 {
		t.Errorf("Expected valid certificate chain")
	}

	// Verify files exist on disk
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("Cert file was not created: %s", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Key file was not created: %s", keyPath)
	}

	// Ensure calling again loads existing without re-generating
	_, err = EnsureCertificates(certPath, keyPath)
	if err != nil {
		t.Errorf("Loading existing certificates failed: %v", err)
	}
}

func TestEnsureCertificatesSelfHealing(t *testing.T) {
	// Case 1: Empty 0-byte cert file is self-healed
	t.Run("SelfHealZeroByteCert", func(t *testing.T) {
		tmpDir := t.TempDir()
		certPath := filepath.Join(tmpDir, "certs", "empty.crt")
		keyPath := filepath.Join(tmpDir, "certs", "empty.key")
		_ = os.MkdirAll(filepath.Dir(certPath), 0755)
		_ = os.WriteFile(certPath, []byte(""), 0644)
		_ = os.WriteFile(keyPath, []byte(""), 0600)

		cert, err := EnsureCertificates(certPath, keyPath)
		if err != nil {
			t.Fatalf("Expected self-healing on 0-byte cert files, got: %v", err)
		}
		if len(cert.Certificate) == 0 {
			t.Errorf("Expected valid cert chain after self-healing")
		}
	})

	// Case 2: Corrupted non-PEM garbage is self-healed
	t.Run("SelfHealCorruptedCert", func(t *testing.T) {
		tmpDir := t.TempDir()
		certPath := filepath.Join(tmpDir, "certs", "corrupt.crt")
		keyPath := filepath.Join(tmpDir, "certs", "corrupt.key")
		_ = os.MkdirAll(filepath.Dir(certPath), 0755)
		_ = os.WriteFile(certPath, []byte("THIS IS NOT A VALID PEM CERTIFICATE"), 0644)
		_ = os.WriteFile(keyPath, []byte("THIS IS NOT A VALID PEM KEY"), 0600)

		cert, err := EnsureCertificates(certPath, keyPath)
		if err != nil {
			t.Fatalf("Expected self-healing on corrupted cert files, got: %v", err)
		}
		if len(cert.Certificate) == 0 {
			t.Errorf("Expected valid cert chain after self-healing corrupted files")
		}
	})
}
