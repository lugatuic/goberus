package handlers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lugatuic/goberus/ldaps"
)

// SanitizeUser trims fields and validates username.
func SanitizeUser(u *ldaps.UserInfo) error {
	if u == nil {
		return fmt.Errorf("nil user")
	}
	// Trim whitespace
	u.Username = strings.TrimSpace(u.Username)
	u.GivenName = strings.TrimSpace(u.GivenName)
	u.Surname = strings.TrimSpace(u.Surname)
	u.DisplayName = strings.TrimSpace(u.DisplayName)
	u.Mail = strings.TrimSpace(u.Mail)
	u.Phone = strings.TrimSpace(u.Phone)
	u.Description = strings.TrimSpace(u.Description)
	u.OrganizationalUnit = strings.TrimSpace(u.OrganizationalUnit)

	if u.Username == "" {
		return fmt.Errorf("username is required")
	}

	var validUser = regexp.MustCompile(`^[A-Za-z0-9@._-]{2,64}$`)
	if !validUser.MatchString(u.Username) {
		return fmt.Errorf("username must be 2-64 characters and contain only letters, numbers, @, ., _, or -")
	}

	u.Username = strings.ToLower(u.Username)
	u.OrganizationalUnit = strings.ToLower(u.OrganizationalUnit)
	return nil
}
