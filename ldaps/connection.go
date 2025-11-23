package ldaps

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/go-ldap/ldap/v3"
	"go.uber.org/zap"

	"github.com/lugatuic/goberus/config"
)

// Client holds configuration and TLS settings for LDAPS connections.
type Client struct {
	cfg       *config.Config
	tlsConfig *tls.Config
	logger    *zap.Logger
}

// NewClient prepares a Client and TLS settings (but does not connect yet).
func NewClient(cfg *config.Config) (*Client, error) {
	c := &Client{cfg: cfg}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	if cfg.CACertPath != "" {
		pool, err := config.LoadCAPool(cfg.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("load CA pool: %w", err)
		}
		if pool != nil {
			tlsCfg.RootCAs = pool
		}
	}

	c.tlsConfig = tlsCfg
	return c, nil
}

func (c *Client) dialAndBind(ctx context.Context) (*ldap.Conn, error) {
	ldapsURL := fmt.Sprintf("ldaps://%s", c.cfg.LdapAddr)
	dialer := &net.Dialer{}
	conn, err := ldap.DialURL(ldapsURL, ldap.DialWithDialer(dialer), ldap.DialWithTLSConfig(c.tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to dial LDAPS %s: %w", ldapsURL, err)
	}

	if dl, ok := ctx.Deadline(); ok {
		conn.SetTimeout(time.Until(dl))
	} else {
		conn.SetTimeout(10 * time.Second)
	}

	if c.cfg.BindDN != "" {
		if bindErr := conn.Bind(c.cfg.BindDN, c.cfg.BindPassword); bindErr != nil {
			conn.Close()
			return nil, fmt.Errorf("service bind failed: %w", bindErr)
		}
	}
	return conn, nil
}
