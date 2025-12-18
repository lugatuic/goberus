package config

import (
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	cfg.SkipVerify = boolFromEnv("LDAP_SKIP_VERIFY", false)
	// CA cert path is optional; if provided, it will be validated at connection time.
	// Do not check existence here to support containers where the CA file may not be
	// available immediately at startup (e.g., Samba initialization in docker-compose).
	// Basic validation
	if cfg.BaseDN == "" {
		return nil, fmt.Errorf("LDAP_BASE_DN must be set")
	}
	return cfg, nil
}

func boolFromEnv(key string, def bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return def
	}
	trimmed := strings.Trim(val, "\"'")
	b, err := strconv.ParseBool(trimmed)
	if err != nil {
		return def
	}
	return b
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
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(pemData); !ok {
		return nil, fmt.Errorf("failed to parse CA certificate(s) from %s", caPath)
	}
	return pool, nil
}
