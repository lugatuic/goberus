package ldaps

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/lugatuic/goberus/config"
	"go.uber.org/zap"
)

// MemberInfo is a minimal struct representing attributes returned by get_member_info.
type MemberInfo struct {
	DN              string   `json:"distinguishedName,omitempty"`
	CN              string   `json:"cn,omitempty"`
	DisplayName     string   `json:"displayName,omitempty"`
	Mail            string   `json:"mail,omitempty"`
	SAMAccountName  string   `json:"sAMAccountName,omitempty"`
	MemberOf        []string `json:"memberOf,omitempty"`
	Description     string   `json:"description,omitempty"`
	BadPasswordTime string   `json:"badPasswordTime,omitempty"`
}

// UserInfo represents the minimal user registration payload used by AddUser.
type UserInfo struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	GivenName   string `json:"givenName,omitempty"`
	Surname     string `json:"surname,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Mail        string `json:"mail,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Major       string `json:"major,omitempty"`
	College     string `json:"college,omitempty"`
	Description string `json:"description,omitempty"`
}

// Client holds configuration and TLS config for dialing LDAPS.
// For simplicity each operation dials/binds and closes the connection.
type Client struct {
	cfg       *config.Config
	tlsConfig *tls.Config
	logger    *zap.Logger
}

// NewClient prepares a Client and TLS settings (but does not connect yet).
func NewClient(cfg *config.Config, logger *zap.Logger) (*Client, error) {
	var c *Client = &Client{cfg: cfg, logger: logger}

	// Build tls.Config
	var tlsCfg *tls.Config = &tls.Config{
		InsecureSkipVerify: cfg.SkipVerify, // if true, skip verification (not recommended)
		MinVersion:         tls.VersionTLS12,
	}

	// If a CA cert is provided, load it and set RootCAs
	if cfg.CACertPath != "" {
		var pool *x509.CertPool
		var err error
		pool, err = config.LoadCAPool(cfg.CACertPath)
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

// helper to dial an LDAPS connection and bind with a service account (if configured).
// The returned conn must be Closed by the caller.
func (c *Client) dialAndBind(ctx context.Context) (*ldap.Conn, error) {
	// Build ldaps URL via DialURL; ldap.DialURL accepts scheme
	// Expect cfg.LdapAddr like "dc.example.local:636" or "1.2.3.4:636"
	var ldapsURL string = fmt.Sprintf("ldaps://%s", c.cfg.LdapAddr)

	// Dial with timeout via context
	var dialer *net.Dialer = &net.Dialer{}
	var conn *ldap.Conn
	var err error
	conn, err = ldap.DialURL(ldapsURL, ldap.DialWithDialer(dialer), ldap.DialWithTLSConfig(c.tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to dial LDAPS %s: %w", ldapsURL, err)
	}

	// Set a per-connection deadline using context deadline if present
	var dl time.Time
	var ok bool
	dl, ok = ctx.Deadline()
	if ok {
		conn.SetTimeout(time.Until(dl))
	} else {
		conn.SetTimeout(10 * time.Second)
	}

	// If a service bind DN is provided, bind now
	if c.cfg.BindDN != "" {
		var bindErr error
		bindErr = conn.Bind(c.cfg.BindDN, c.cfg.BindPassword)
		if bindErr != nil {
			if c.logger != nil {
				c.logger.Error("service bind failed", zap.Error(bindErr), zap.String("bind_dn", c.cfg.BindDN))
			}
			conn.Close()
			return nil, fmt.Errorf("service bind failed: %w", bindErr)
		}
	}
	return conn, nil
}

// GetMemberInfo searches for a user by userPrincipalName or sAMAccountName and returns selected attributes.
func (c *Client) GetMemberInfo(ctx context.Context, username string) (*MemberInfo, error) {
	// create a sub-context with timeout
	var ctxWithTimeout context.Context
	var cancel context.CancelFunc
	ctxWithTimeout, cancel = context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	conn, err := c.dialAndBind(ctxWithTimeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// build filter to accept either UPN or sAMAccountName
	// Escape input to avoid LDAP injection - one-line explicit typed declarations
	var esc string = ldap.EscapeFilter(username)
	var filter string = fmt.Sprintf("(|(userPrincipalName=%s)(sAMAccountName=%s))", esc, esc)

	var attributes []string = []string{
		"distinguishedName",
		"cn",
		"displayName",
		"mail",
		"sAMAccountName",
		"memberOf",
		"description",
		"badPasswordTime",
	}

	var searchReq *ldap.SearchRequest = ldap.NewSearchRequest(
		c.cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,  // size limit - we only expect one user
		10, // time limit in seconds
		false,
		filter,
		attributes,
		nil,
	)

	// Perform search; note: ldap package does not accept context, so we use conn.SetTimeout above.
	var sr *ldap.SearchResult
	sr, err = conn.Search(searchReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ldap search failed", zap.Error(err), zap.String("filter", filter), zap.String("username", username))
		}
		return nil, fmt.Errorf("ldap search failed: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("no entries found for %s", username)
	}
	var entry *ldap.Entry
	entry = sr.Entries[0]

	var info *MemberInfo = &MemberInfo{
		DN:              entry.GetAttributeValue("distinguishedName"),
		CN:              entry.GetAttributeValue("cn"),
		DisplayName:     entry.GetAttributeValue("displayName"),
		Mail:            entry.GetAttributeValue("mail"),
		SAMAccountName:  entry.GetAttributeValue("sAMAccountName"),
		Description:     entry.GetAttributeValue("description"),
		BadPasswordTime: entry.GetAttributeValue("badPasswordTime"),
	}

	// memberOf can be multi-valued
	var members []string = entry.GetAttributeValues("memberOf")
	if len(members) > 0 {
		var normalized []string = make([]string, 0, len(members))
		var i int
		for i = 0; i < len(members); i++ {
			var m string = members[i]
			normalized = append(normalized, strings.TrimSpace(m))
		}
		info.MemberOf = normalized
	}

	return info, nil
}

// AddUser creates a new LDAP entry for the provided user info.
func (c *Client) AddUser(ctx context.Context, u *UserInfo) error {
	// small timeout for add
	var ctxWithTimeout context.Context
	var cancel context.CancelFunc
	ctxWithTimeout, cancel = context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	conn, err := c.dialAndBind(ctxWithTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Build DN: if UserOU provided in config, include it
	var dn string
	if c.cfg.UserOU != "" {
		dn = fmt.Sprintf("CN=%s,%s,%s", ldap.EscapeFilter(u.Username), c.cfg.UserOU, c.cfg.BaseDN)
	} else {
		dn = fmt.Sprintf("CN=%s,%s", ldap.EscapeFilter(u.Username), c.cfg.BaseDN)
	}

	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})

	// Basic attributes
	req.Attribute("cn", []string{u.Username})
	if u.Surname != "" {
		req.Attribute("sn", []string{u.Surname})
	}
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

	// Try to build a sensible userPrincipalName from BaseDN DC components
	var upn string
	parts := strings.Split(c.cfg.BaseDN, ",")
	var dcs []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(strings.ToLower(p), "dc=") {
			dcs = append(dcs, strings.TrimSpace(p[3:]))
		}
	}
	if len(dcs) > 0 {
		upn = fmt.Sprintf("%s@%s", u.Username, strings.Join(dcs, "."))
		req.Attribute("userPrincipalName", []string{upn})
	}

	// Password: AD may require specific attribute handling; we set userPassword for now.
	if u.Password != "" {
		req.Attribute("userPassword", []string{u.Password})
	}

	if err := conn.Add(req); err != nil {
		if c.logger != nil {
			c.logger.Error("ldap add failed", zap.Error(err), zap.String("dn", dn), zap.String("username", u.Username))
		}
		return fmt.Errorf("ldap add failed: %w", err)
	}

	if c.logger != nil {
		c.logger.Info("user added", zap.String("dn", dn), zap.String("username", u.Username))
	}

	return nil
}
