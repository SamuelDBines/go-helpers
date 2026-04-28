package certs

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	CertFile     string
	KeyFile      string
	Hosts        []string
	CommonName   string
	Organization string
	ValidFor     time.Duration
}

func LoadOrCreate(cfg Config) (tls.Certificate, error) {
	cfg = withDefaults(cfg)

	if cfg.CertFile != "" || cfg.KeyFile != "" {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return tls.Certificate{}, errors.New("certs: cert and key files must both be set")
		}

		certExists, err := fileExists(cfg.CertFile)
		if err != nil {
			return tls.Certificate{}, err
		}
		keyExists, err := fileExists(cfg.KeyFile)
		if err != nil {
			return tls.Certificate{}, err
		}

		switch {
		case certExists && keyExists:
			return tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		case certExists != keyExists:
			return tls.Certificate{}, fmt.Errorf("certs: only one of cert or key file exists (%s, %s)", cfg.CertFile, cfg.KeyFile)
		}
	}

	certPEM, keyPEM, err := GenerateSelfSignedPEM(cfg)
	if err != nil {
		return tls.Certificate{}, err
	}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		if err := writeKeyPair(cfg.CertFile, cfg.KeyFile, certPEM, keyPEM); err != nil {
			return tls.Certificate{}, err
		}
	}

	return tls.X509KeyPair(certPEM, keyPEM)
}

func ServerTLSConfig(cfg Config) (*tls.Config, error) {
	cert, err := LoadOrCreate(cfg)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2", "http/1.1"},
	}, nil
}

func GenerateSelfSignedPEM(cfg Config) ([]byte, []byte, error) {
	cfg = withDefaults(cfg)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("certs: generate key: %w", err)
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("certs: generate serial number: %w", err)
	}

	tpl := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: []string{cfg.Organization},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(cfg.ValidFor),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames(cfg.Hosts),
		IPAddresses:           ipAddresses(cfg.Hosts),
	}

	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("certs: create certificate: %w", err)
	}

	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("certs: marshal private key: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	return certPEM, keyPEM, nil
}

func withDefaults(cfg Config) Config {
	if len(cfg.Hosts) == 0 {
		cfg.Hosts = []string{"localhost", "127.0.0.1", "::1"}
	}
	if cfg.ValidFor == 0 {
		cfg.ValidFor = 365 * 24 * time.Hour
	}
	if cfg.Organization == "" {
		cfg.Organization = "go-helpers"
	}
	if cfg.CommonName == "" {
		cfg.CommonName = firstUsableName(cfg.Hosts)
	}
	return cfg
}

func writeKeyPair(certFile string, keyFile string, certPEM []byte, keyPEM []byte) error {
	if err := os.MkdirAll(filepath.Dir(certFile), 0o755); err != nil {
		return fmt.Errorf("certs: create cert directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyFile), 0o755); err != nil {
		return fmt.Errorf("certs: create key directory: %w", err)
	}

	if err := os.WriteFile(certFile, certPEM, 0o644); err != nil {
		return fmt.Errorf("certs: write cert file: %w", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		return fmt.Errorf("certs: write key file: %w", err)
	}
	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("certs: stat %s: %w", path, err)
}

func dnsNames(hosts []string) []string {
	var names []string
	for _, host := range hosts {
		host = cleanHost(host)
		if host == "" {
			continue
		}
		if net.ParseIP(host) != nil {
			continue
		}
		names = append(names, host)
	}
	return unique(names)
}

func ipAddresses(hosts []string) []net.IP {
	var ips []net.IP
	for _, host := range hosts {
		host = cleanHost(host)
		if host == "" {
			continue
		}
		if ip := net.ParseIP(host); ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips
}

func firstUsableName(hosts []string) string {
	for _, host := range hosts {
		host = cleanHost(host)
		if host != "" {
			return host
		}
	}
	return "localhost"
}

func cleanHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	return host
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
