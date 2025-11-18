package config

import (
	"crypto/x509"
	"fmt"
	"os"
)

type Config struct {
	BindAddr     string // HTTP bind address, e.g. :8080
	LdapAddr     string // host:port, e.g. dc.example.local:636
	BaseDN       string
	BindDN       string
	BindPassword string
	SkipVerify   bool
	CACertPath   string // optional path to CA PEM to verify LDAPS certs
}

func LoadFromEnv() (*Config, error) {
	var cfg *Config = &Config{
		BindAddr:     getenv("BIND_ADDR", ":8080"),
		LdapAddr:     getenv("LDAP_ADDR", "dc.example.local:636"),
		BaseDN:       getenv("LDAP_BASE_DN", "dc=example,dc=local"),
		BindDN:       os.Getenv("LDAP_BIND_DN"),
		BindPassword: os.Getenv("LDAP_BIND_PASSWORD"),
		CACertPath:   os.Getenv("LDAP_CA_CERT"),
	}
	if getenv("LDAP_SKIP_VERIFY", "false") == "true" {
		cfg.SkipVerify = true
	}
	// If user provided a CA cert path, ensure it exists (we'll try to load it later)
	if cfg.CACertPath != "" {
		var statErr error
		_, statErr = os.Stat(cfg.CACertPath)
		if statErr != nil {
			return nil, fmt.Errorf("LDAP_CA_CERT file not accessible: %w", statErr)
		}
	}
	// Basic validation
	if cfg.BaseDN == "" {
		return nil, fmt.Errorf("LDAP_BASE_DN must be set")
	}
	return cfg, nil
}

func getenv(k, def string) string {
	var v string = os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

// helper to load CA pool — callers can use this to build tls.Config
func LoadCAPool(caPath string) (*x509.CertPool, error) {
	if caPath == "" {
		return nil, nil
	}
	pemData, err := os.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	var pool *x509.CertPool = x509.NewCertPool()
	var ok bool
	ok = pool.AppendCertsFromPEM(pemData)
	if !ok {
		return nil, fmt.Errorf("failed to parse CA certificate(s) from %s", caPath)
	}
	return pool, nil
}