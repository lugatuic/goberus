package ldaps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"go.uber.org/zap"
)

// GetMemberInfo searches for a user by userPrincipalName or sAMAccountName and returns selected attributes.
func (c *Client) GetMemberInfo(ctx context.Context, username string) (*MemberInfo, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	conn, err := c.dialAndBind(ctxWithTimeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

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
		1,
		10,
		false,
		filter,
		attributes,
		nil,
	)

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
