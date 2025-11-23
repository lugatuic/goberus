package ldaps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"go.uber.org/zap"
)

// AddUser creates a new LDAP entry for the provided user information.
func (c *Client) AddUser(ctx context.Context, u *UserInfo) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	conn, err := c.dialAndBind(ctxTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	dn := c.buildUserDN(u)
	req := c.buildAddRequest(dn, u)

	if err := conn.Add(req); err != nil {
		if c.logger != nil {
			c.logger.Error("ldap add failed", zap.Error(err), zap.String("dn", dn), zap.String("username", u.Username))
		}
		return fmt.Errorf("ldap add failed: %w", err)
	}

	if u.Password != "" {
		if err := c.setUnicodePwd(conn, dn, u.Password); err != nil {
			if c.logger != nil {
				c.logger.Error("set unicodePwd failed", zap.Error(err), zap.String("dn", dn), zap.String("username", u.Username))
			}
			return err
		}
		if err := c.enableAccount(conn, dn); err != nil {
			if c.logger != nil {
				c.logger.Warn("enable account failed", zap.Error(err), zap.String("dn", dn), zap.String("username", u.Username))
			}
			return err
		}
	}

	if c.logger != nil {
		c.logger.Info("user added", zap.String("dn", dn), zap.String("username", u.Username))
	}

	return nil
}

func (c *Client) buildAddRequest(dn string, u *UserInfo) *ldap.AddRequest {
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
	req.Attribute("cn", []string{u.Username})

	sn := u.Surname
	if sn == "" {
		sn = u.Username
	}
	req.Attribute("sn", []string{sn})

	if u.DisplayName != "" {
		req.Attribute("displayName", []string{u.DisplayName})
	}
	req.Attribute("sAMAccountName", []string{u.Username})

	if u.Mail != "" {
		req.Attribute("mail", []string{u.Mail})
	}
	if u.Phone != "" {
		req.Attribute("telephoneNumber", []string{u.Phone})
	}
	if u.Description != "" {
		req.Attribute("description", []string{u.Description})
	}

	dcParts := parseDCParts(c.cfg.BaseDN)
	if len(dcParts) > 0 {
		domain := strings.Join(dcParts, ".")
		upn := fmt.Sprintf("%s@%s", u.Username, domain)
		req.Attribute("userPrincipalName", []string{upn})
	}

	return req
}

func parseDCParts(baseDN string) []string {
	parts := strings.Split(baseDN, ",")
	var dcParts []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(strings.ToLower(p), "dc=") {
			if len(p) > 3 {
				dcParts = append(dcParts, p[3:])
			}
		}
	}
	return dcParts
}
