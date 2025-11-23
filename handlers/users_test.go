package handlers

import (
	"strings"
	"testing"

	"github.com/matryer/is"

	"github.com/lugatuic/goberus/ldaps"
)

func TestSanitizeUser(t *testing.T) {
	t.Run("trims and normalizes", func(t *testing.T) {
		is := is.New(t)
		user := &ldaps.UserInfo{
			Username:    " ExampleUser ",
			Password:    " secret ",
			GivenName:   " John ",
			Surname:     " Doe ",
			Mail:        " test@example.com ",
			Phone:       " 555-1234 ",
			Description: " Desc ",
		}

		is.NoErr(SanitizeUser(user))
		is.Equal(user.Username, "exampleuser")
		is.Equal(user.Password, "secret")
		is.Equal(user.GivenName, "John")
		is.Equal(user.Surname, "Doe")
		is.Equal(user.Mail, "test@example.com")
		is.Equal(user.Phone, "555-1234")
		is.Equal(user.Description, "Desc")
	})

	t.Run("rejects invalid username", func(t *testing.T) {
		is := is.New(t)

		err := SanitizeUser(&ldaps.UserInfo{Username: "!"})
		is.True(err != nil)
		is.True(strings.Contains(err.Error(), "invalid characters"))
	})
}
