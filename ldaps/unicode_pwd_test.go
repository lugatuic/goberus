package ldaps

import (
	"errors"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/matryer/is"
)

type mockModifier struct {
	lastRequest *ldap.ModifyRequest
	err         error
	calls       int
}

func (m *mockModifier) Modify(req *ldap.ModifyRequest) error {
	m.calls++
	m.lastRequest = req
	if m.err != nil {
		return m.err
	}
	return nil
}

func TestEncodeUnicodePwd(t *testing.T) {
	is := is.New(t)
	got := encodeUnicodePwd("Secret123")
	want := []byte{34, 0, 83, 0, 101, 0, 99, 0, 114, 0, 101, 0, 116, 0, 49, 0, 50, 0, 51, 0, 34, 0}
	is.Equal(got, want)
}

func TestSetUnicodePwd(t *testing.T) {
	t.Run("no password skips modify", func(t *testing.T) {
		is := is.New(t)
		modifier := &mockModifier{}
		client := &Client{}
		is.NoErr(client.setUnicodePwd(modifier, "cn=foo", ""))
		is.Equal(modifier.calls, 0)
	})

	t.Run("success writes unicodePwd", func(t *testing.T) {
		is := is.New(t)
		modifier := &mockModifier{}
		client := &Client{}
		password := "Passw0rd!"
		is.NoErr(client.setUnicodePwd(modifier, "cn=user", password))
		is.Equal(modifier.calls, 1)
		is.True(modifier.lastRequest != nil)
		is.Equal(modifier.lastRequest.DN, "cn=user")
		changes := modifier.lastRequest.Changes
		is.Equal(len(changes), 1)
		is.Equal(changes[0].Modification.Type, "unicodePwd")
		is.Equal([]byte(changes[0].Modification.Vals[0]), encodeUnicodePwd(password))
	})

	t.Run("propagates modify error", func(t *testing.T) {
		is := is.New(t)
		modifier := &mockModifier{err: errors.New("boom")}
		client := &Client{}
		err := client.setUnicodePwd(modifier, "cn=user", "pw")
		is.True(err != nil)
		is.Equal(modifier.calls, 1)
	})
}

func TestEnableAccount(t *testing.T) {
	t.Run("success writes control", func(t *testing.T) {
		is := is.New(t)
		modifier := &mockModifier{}
		client := &Client{}
		is.NoErr(client.enableAccount(modifier, "cn=user"))
		is.Equal(modifier.calls, 1)
		is.Equal(modifier.lastRequest.DN, "cn=user")
		changes := modifier.lastRequest.Changes
		is.Equal(changes[0].Modification.Type, "userAccountControl")
		is.Equal(changes[0].Modification.Vals, []string{"512"})
	})

	t.Run("returns modify error", func(t *testing.T) {
		is := is.New(t)
		modifier := &mockModifier{err: errors.New("failed")}
		client := &Client{}
		err := client.enableAccount(modifier, "cn=user")
		is.True(err != nil)
		is.Equal(modifier.calls, 1)
	})
}
