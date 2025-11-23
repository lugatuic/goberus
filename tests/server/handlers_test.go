package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lugatuic/goberus/ldaps"
	"github.com/lugatuic/goberus/server"
	"github.com/matryer/is"
)

type fakeUserClient struct {
	getMemberInfo func(ctx context.Context, username string) (*ldaps.MemberInfo, error)
	addUser       func(ctx context.Context, u *ldaps.UserInfo) error
}

var _ server.UserClient = (*fakeUserClient)(nil)

func (f *fakeUserClient) GetMemberInfo(ctx context.Context, username string) (*ldaps.MemberInfo, error) {
	if f.getMemberInfo != nil {
		return f.getMemberInfo(ctx, username)
	}
	return nil, errors.New("GetMemberInfo not stubbed")
}

func (f *fakeUserClient) AddUser(ctx context.Context, u *ldaps.UserInfo) error {
	if f.addUser != nil {
		return f.addUser(ctx, u)
	}
	return nil
}

func TestHandleGetMember(t *testing.T) {
	t.Run("missing username", func(t *testing.T) {
		is := is.New(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/member", nil)
		rr := httptest.NewRecorder()

		is.NoErr(server.HandleGetMember(&fakeUserClient{}, rr, req))
		is.Equal(rr.Code, http.StatusBadRequest)
		is.True(strings.Contains(rr.Body.String(), "missing username parameter"))
	})

	t.Run("client error", func(t *testing.T) {
		is := is.New(t)
		wantErr := errors.New("boom")
		client := &fakeUserClient{
			getMemberInfo: func(ctx context.Context, username string) (*ldaps.MemberInfo, error) {
				return nil, wantErr
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/v1/member?username=jdoe", nil)
		rr := httptest.NewRecorder()

		is.Equal(server.HandleGetMember(client, rr, req), wantErr)
	})

	t.Run("success", func(t *testing.T) {
		is := is.New(t)
		want := &ldaps.MemberInfo{DisplayName: "Jane"}
		client := &fakeUserClient{
			getMemberInfo: func(ctx context.Context, username string) (*ldaps.MemberInfo, error) {
				is.Equal(username, "jdoe")
				return want, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/v1/member?username=jdoe", nil)
		rr := httptest.NewRecorder()

		is.NoErr(server.HandleGetMember(client, rr, req))
		is.Equal(rr.Code, http.StatusOK)
		is.Equal(rr.Header().Get("Content-Type"), "application/json")

		var got ldaps.MemberInfo
		is.NoErr(json.Unmarshal(rr.Body.Bytes(), &got))
		is.Equal(got.DisplayName, want.DisplayName)
	})
}

func TestHandleCreateMember(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		is := is.New(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/member", strings.NewReader("{not json"))
		rr := httptest.NewRecorder()

		is.NoErr(server.HandleCreateMember(&fakeUserClient{}, rr, req))
		is.Equal(rr.Code, http.StatusBadRequest)
		is.True(strings.Contains(rr.Body.String(), "invalid json"))
	})

	t.Run("invalid input", func(t *testing.T) {
		is := is.New(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/member", strings.NewReader(`{"username":"a"}`))
		rr := httptest.NewRecorder()

		is.NoErr(server.HandleCreateMember(&fakeUserClient{}, rr, req))
		is.Equal(rr.Code, http.StatusBadRequest)
		is.True(strings.Contains(rr.Body.String(), "invalid input"))
	})

	t.Run("success", func(t *testing.T) {
		is := is.New(t)
		var captured *ldaps.UserInfo
		client := &fakeUserClient{
			addUser: func(ctx context.Context, u *ldaps.UserInfo) error {
				captured = u
				return nil
			},
		}
		body := strings.NewReader(`{"username":" TestUser ","password":"secret"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/member", body)
		rr := httptest.NewRecorder()

		is.NoErr(server.HandleCreateMember(client, rr, req))
		is.Equal(rr.Code, http.StatusCreated)
		is.Equal(rr.Header().Get("Content-Type"), "application/json")
		is.Equal(rr.Body.String(), "{\"status\":\"created\"}\n")
		is.Equal(captured.Username, "testuser")
	})
}

func TestSanitizeUserIntegration(t *testing.T) {
	is := is.New(t)
	var captured *ldaps.UserInfo
	client := &fakeUserClient{
		addUser: func(ctx context.Context, u *ldaps.UserInfo) error {
			captured = u
			return nil
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/member", func(w http.ResponseWriter, r *http.Request) {
		is.NoErr(server.HandleCreateMember(client, w, r))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/v1/member", "application/json", strings.NewReader(`{"username":" TestUser ","password":" secret ","ou":" OU=ACMUsers,DC=acmuic,DC=org "}`))
	is.NoErr(err)
	defer resp.Body.Close()
	is.Equal(resp.StatusCode, http.StatusCreated)
	is.Equal(captured.Username, "testuser")
	is.Equal(captured.OrganizationalUnit, "ou=acmusers,dc=acmuic,dc=org")
}
