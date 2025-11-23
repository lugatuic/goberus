package ldaps

import (
	"fmt"
	"strings"
)

func escapeDNComponent(s string) string {
	var builder strings.Builder
	for i, r := range s {
		isSpecial := false
		switch r {
		case '\\', ',', '+', '"', '<', '>', ';', '#', '=':
			isSpecial = true
		}

		if isSpecial || (i == 0 && r == ' ') || (i == len(s)-1 && r == ' ') {
			builder.WriteRune('\\')
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func (c *Client) buildUserDN(u *UserInfo) string {
	escCN := escapeDNComponent(u.Username)
	ou := strings.TrimSpace(u.OrganizationalUnit)
	if ou != "" {
		if strings.HasSuffix(strings.ToLower(ou), strings.ToLower(c.cfg.BaseDN)) {
			return fmt.Sprintf("CN=%s,%s", escCN, ou)
		}
		return fmt.Sprintf("CN=%s,%s,%s", escCN, ou, c.cfg.BaseDN)
	}
	return fmt.Sprintf("CN=%s,%s", escCN, c.cfg.BaseDN)
}
