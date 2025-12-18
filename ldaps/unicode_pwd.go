package ldaps

import (
	"fmt"
	"unicode/utf16"

	"github.com/go-ldap/ldap/v3"
)

type ldapModifier interface {
	Modify(*ldap.ModifyRequest) error
}

func (c *Client) setUnicodePwd(conn ldapModifier, dn, password string) error {
	if password == "" {
		return nil
	}
	pwdBytes := encodeUnicodePwd(password)
	mr := ldap.NewModifyRequest(dn, nil)
	mr.Replace("unicodePwd", []string{string(pwdBytes)})
	if err := conn.Modify(mr); err != nil {
		return fmt.Errorf("set unicodePwd failed: %w", err)
	}
	return nil
}

func (c *Client) enableAccount(conn ldapModifier, dn string) error {
	mr := ldap.NewModifyRequest(dn, nil)
	mr.Replace("userAccountControl", []string{"512"})
	if err := conn.Modify(mr); err != nil {
		return fmt.Errorf("enable account failed: %w", err)
	}
	return nil
}

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
