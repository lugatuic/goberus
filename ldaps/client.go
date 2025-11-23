package ldaps

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/go-ldap/ldap/v3"
	"go.uber.org/zap"

	"github.com/lugatuic/goberus/config"
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

// (previously exported a sentinel ErrEntryExists to allow mapping to HTTP 409)

// NewClient prepares a Client and TLS settings (but does not connect yet).
func NewClient(cfg *config.Config) (*Client, error) {
	c := &Client{cfg: cfg}

	// Build tls.Config
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerify, // if true, skip verification (not recommended)
		MinVersion:         tls.VersionTLS12,
	}

	// If a CA cert is provided, load it and set RootCAs
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

// helper to dial an LDAPS connection and bind with a service account (if configured).
// The returned conn must be Closed by the caller.
func (c *Client) dialAndBind(ctx context.Context) (*ldap.Conn, error) {
	// Build ldaps URL via DialURL; ldap.DialURL accepts scheme
	// Expect cfg.LdapAddr like "dc.example.local:636" or "1.2.3.4:636"
	ldapsURL := fmt.Sprintf("ldaps://%s", c.cfg.LdapAddr)

	// Dial with timeout via context
	dialer := &net.Dialer{}
	conn, err := ldap.DialURL(ldapsURL, ldap.DialWithDialer(dialer), ldap.DialWithTLSConfig(c.tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to dial LDAPS %s: %w", ldapsURL, err)
	}

	// Set a per-connection deadline using context deadline if present
	if dl, ok := ctx.Deadline(); ok {
		conn.SetTimeout(time.Until(dl))
	} else {
		conn.SetTimeout(10 * time.Second)
	}

	// If a service bind DN is provided, bind now
	if c.cfg.BindDN != "" {
		if bindErr := conn.Bind(c.cfg.BindDN, c.cfg.BindPassword); bindErr != nil {
			conn.Close()
			return nil, fmt.Errorf("service bind failed: %w", bindErr)
		}
	}
	return conn, nil
}

// GetMemberInfo searches for a user by userPrincipalName or sAMAccountName and returns selected attributes.
func (c *Client) GetMemberInfo(ctx context.Context, username string) (*MemberInfo, error) {
	// create a sub-context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	conn, err := c.dialAndBind(ctxWithTimeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// build filter to accept either UPN or sAMAccountName
	// Escape input to avoid LDAP injection - one-line explicit typed declarations
	esc := ldap.EscapeFilter(username)
	filter := fmt.Sprintf("(|(userPrincipalName=%s)(sAMAccountName=%s))", esc, esc)

	attributes := []string{
		"distinguishedName",
		"cn",
		"displayName",
		"mail",
		"sAMAccountName",
		"memberOf",
		"description",
		"badPasswordTime",
	}

	searchReq := ldap.NewSearchRequest(
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
	sr, err := conn.Search(searchReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ldap search failed", zap.Error(err), zap.String("filter", filter), zap.String("username", username))
		}
		return nil, fmt.Errorf("ldap search failed: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("no entries found for %s", username)
	}
	entry := sr.Entries[0]

	info := &MemberInfo{
		DN:              entry.GetAttributeValue("distinguishedName"),
		CN:              entry.GetAttributeValue("cn"),
		DisplayName:     entry.GetAttributeValue("displayName"),
		Mail:            entry.GetAttributeValue("mail"),
		SAMAccountName:  entry.GetAttributeValue("sAMAccountName"),
		Description:     entry.GetAttributeValue("description"),
		BadPasswordTime: entry.GetAttributeValue("badPasswordTime"),
	}

	// memberOf can be multi-valued
	members := entry.GetAttributeValues("memberOf")
	if len(members) > 0 {
		normalized := make([]string, 0, len(members))
		for _, m := range members {
			normalized = append(normalized, strings.TrimSpace(m))
		}
		info.MemberOf = normalized
	}

	return info, nil
}

// escapeDNComponent escapes common DN-special characters.
func escapeDNComponent(s string) string {
	var builder strings.Builder
	for i, r := range s {
		isSpecial := false
		switch r {
		case '\\', ',', '+', '"', '<', '>', ';':
			isSpecial = true
		}

		if isSpecial || (i == 0 && r == ' ') || (i == len(s)-1 && r == ' ') {
			builder.WriteRune('\\')
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

// AddUser creates a new LDAP entry for the provided user information.
func (c *Client) AddUser(ctx context.Context, u *UserInfo) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	conn, err := c.dialAndBind(ctxTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Build DN: CN=<escaped username>,<UserOU>,<BaseDN> or CN=<escaped username>,<BaseDN>
	escCN := escapeDNComponent(u.Username)
	var dn string
	if c.cfg.UserOU != "" {
		dn = fmt.Sprintf("CN=%s,%s,%s", escCN, c.cfg.UserOU, c.cfg.BaseDN)
	} else {
		dn = fmt.Sprintf("CN=%s,%s", escCN, c.cfg.BaseDN)
	}

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

	// userPrincipalName: construct from BaseDN's dc components if possible
	// e.g. BaseDN dc=example,dc=local -> example.local
	parts := strings.Split(c.cfg.BaseDN, ",")
	var dcParts []string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(strings.ToLower(p), "dc=") {
			dcParts = append(dcParts, strings.TrimPrefix(p[3:], ""))
		}
	}

	if len(dcParts) > 0 {
		domain := strings.Join(dcParts, ".")
		upn := fmt.Sprintf("%s@%s", u.Username, domain)
		req.Attribute("userPrincipalName", []string{upn})
	}

	// For Active Directory, set password via `unicodePwd` (UTF-16LE, quoted),
	// performed as a separate modify after the add.

	if err := conn.Add(req); err != nil {
		if c.logger != nil {
			c.logger.Error("ldap add failed", zap.Error(err), zap.String("dn", dn), zap.String("username", u.Username))
		}
		return fmt.Errorf("ldap add failed: %w", err)
	}

	// If a password was provided, attempt AD-compatible password set via unicodePwd.
	if u.Password != "" {
		pwdBytes := encodeUnicodePwd(u.Password)
		mr := ldap.NewModifyRequest(dn, nil)
		// Replace unicodePwd with the UTF-16LE encoded quoted password. This must
		// be done over LDAPS and the bind account must have permission.
		mr.Replace("unicodePwd", []string{string(pwdBytes)})
		if err := conn.Modify(mr); err != nil {
			if c.logger != nil {
				c.logger.Error("set unicodePwd failed", zap.Error(err), zap.String("dn", dn), zap.String("username", u.Username))
			}
			return fmt.Errorf("set unicodePwd failed: %w", err)
		}

		// Optionally enable the account by setting userAccountControl to 512 (NORMAL_ACCOUNT).
		// Many ADs create accounts disabled by default; enabling requires appropriate rights.
		enableMr := ldap.NewModifyRequest(dn, nil)
		enableMr.Replace("userAccountControl", []string{"512"})
		if err := conn.Modify(enableMr); err != nil {
			if c.logger != nil {
				c.logger.Warn("enable account failed", zap.Error(err), zap.String("dn", dn), zap.String("username", u.Username))
			}
			// Not a hard failure in all environments; surface as an error so calling code can see it.
			return fmt.Errorf("enable account failed: %w", err)
		}
	}

	if c.logger != nil {
		c.logger.Info("user added", zap.String("dn", dn), zap.String("username", u.Username))
	}

	return nil
}

// encodeUnicodePwd encodes the quoted password to UTF-16LE for `unicodePwd`.
func encodeUnicodePwd(pw string) []byte {
	quoted := "\"" + pw + "\""
	u := utf16.Encode([]rune(quoted))
	b := make([]byte, 2*len(u))
	for i, v := range u {
		b[i*2] = byte(v)
		b[i*2+1] = byte(v >> 8)
	}
	return b
}
