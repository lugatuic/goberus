package httpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
	"go.uber.org/zap"

	"github.com/lugatuic/goberus/config"
	"github.com/lugatuic/goberus/internal/httpserver"
	"github.com/lugatuic/goberus/ldaps"
)

type fakeClient struct {
	pingErr       error
	getMemberInfo func(ctx context.Context, username string) (*ldaps.MemberInfo, error)
	addUser       func(ctx context.Context, u *ldaps.UserInfo) error
}

func (f *fakeClient) Ping(ctx context.Context) error {
	return f.pingErr
}

func (f *fakeClient) GetMemberInfo(ctx context.Context, username string) (*ldaps.MemberInfo, error) {
	if f.getMemberInfo != nil {
		return f.getMemberInfo(ctx, username)
	}
	return nil, errors.New("GetMemberInfo not stubbed")
}

func (f *fakeClient) AddUser(ctx context.Context, u *ldaps.UserInfo) error {
	if f.addUser != nil {
		return f.addUser(ctx, u)
	}
	return nil
}

func TestHealthEndpoints(t *testing.T) {
	t.Run("/livez returns OK", func(t *testing.T) {
		is := is.New(t)
		logger := zap.NewNop()
		cfg := &config.Config{BindAddr: ":8080"}
		client := &fakeClient{}

		s := httpserver.New(cfg, logger, client)
		handler := s.Handler()

		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusOK)
		is.Equal(rr.Header().Get("Content-Type"), "application/json; charset=utf-8")

		var resp map[string]string
		is.NoErr(json.Unmarshal(rr.Body.Bytes(), &resp))
		is.Equal(resp["status"], "ok")
	})

	t.Run("/readyz returns OK when LDAP ping succeeds", func(t *testing.T) {
		is := is.New(t)
		logger := zap.NewNop()
		cfg := &config.Config{BindAddr: ":8080"}
		client := &fakeClient{pingErr: nil}

		s := httpserver.New(cfg, logger, client)
		handler := s.Handler()

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusOK)
		is.Equal(rr.Header().Get("Content-Type"), "application/json; charset=utf-8")

		var resp map[string]string
		is.NoErr(json.Unmarshal(rr.Body.Bytes(), &resp))
		is.Equal(resp["status"], "ready")
	})

	t.Run("/readyz returns degraded when LDAP ping fails", func(t *testing.T) {
		is := is.New(t)
		logger := zap.NewNop()
		cfg := &config.Config{BindAddr: ":8080"}
		client := &fakeClient{pingErr: errors.New("connection failed")}

		s := httpserver.New(cfg, logger, client)
		handler := s.Handler()

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusServiceUnavailable)
		is.Equal(rr.Header().Get("Content-Type"), "application/json; charset=utf-8")

		var resp map[string]string
		is.NoErr(json.Unmarshal(rr.Body.Bytes(), &resp))
		is.Equal(resp["status"], "degraded")
	})
}

func TestBusinessRoutes(t *testing.T) {
	t.Run("/v1/member GET success", func(t *testing.T) {
		is := is.New(t)
		logger := zap.NewNop()
		cfg := &config.Config{BindAddr: ":8080"}
		client := &fakeClient{
			getMemberInfo: func(ctx context.Context, username string) (*ldaps.MemberInfo, error) {
				return &ldaps.MemberInfo{DisplayName: "Jane Doe"}, nil
			},
		}

		s := httpserver.New(cfg, logger, client)
		handler := s.Handler()

		req := httptest.NewRequest(http.MethodGet, "/v1/member?username=jdoe", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusOK)
		is.True(rr.Header().Get("X-Request-ID") != "") // RequestID middleware should add this
	})

	t.Run("/v1/member GET error returns sanitized JSON error", func(t *testing.T) {
		is := is.New(t)
		logger := zap.NewNop()
		cfg := &config.Config{BindAddr: ":8080"}
		client := &fakeClient{
			getMemberInfo: func(ctx context.Context, username string) (*ldaps.MemberInfo, error) {
				return nil, errors.New("internal database connection failed with secret details")
			},
		}

		s := httpserver.New(cfg, logger, client)
		handler := s.Handler()

		req := httptest.NewRequest(http.MethodGet, "/v1/member?username=jdoe", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusInternalServerError)
		is.Equal(rr.Header().Get("Content-Type"), "application/json; charset=utf-8")

		var resp map[string]string
		is.NoErr(json.Unmarshal(rr.Body.Bytes(), &resp))
		is.Equal(resp["error"], "internal server error")
		// Ensure internal details are NOT leaked
		is.True(!strings.Contains(rr.Body.String(), "secret details"))
		is.True(!strings.Contains(rr.Body.String(), "database connection"))
	})

	t.Run("/v1/member POST success", func(t *testing.T) {
		is := is.New(t)
		logger := zap.NewNop()
		cfg := &config.Config{BindAddr: ":8080"}
		client := &fakeClient{
			addUser: func(ctx context.Context, u *ldaps.UserInfo) error {
				return nil
			},
		}

		s := httpserver.New(cfg, logger, client)
		handler := s.Handler()

		body := strings.NewReader(`{"username":"testuser","password":"secret"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/member", body)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusCreated)
		is.True(rr.Header().Get("X-Request-ID") != "")
	})

	t.Run("/v1/member unsupported method returns JSON error", func(t *testing.T) {
		is := is.New(t)
		logger := zap.NewNop()
		cfg := &config.Config{BindAddr: ":8080"}
		client := &fakeClient{}

		s := httpserver.New(cfg, logger, client)
		handler := s.Handler()

		req := httptest.NewRequest(http.MethodPut, "/v1/member", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusMethodNotAllowed)
		is.Equal(rr.Header().Get("Content-Type"), "application/json; charset=utf-8")

		var resp map[string]string
		is.NoErr(json.Unmarshal(rr.Body.Bytes(), &resp))
		is.Equal(resp["error"], "method not allowed")
	})
}

func TestRequestIDPreservation(t *testing.T) {
	is := is.New(t)
	logger := zap.NewNop()
	cfg := &config.Config{BindAddr: ":8080"}
	client := &fakeClient{}

	s := httpserver.New(cfg, logger, client)
	handler := s.Handler()

	existingID := "my-custom-request-id"
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	req.Header.Set("X-Request-ID", existingID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	is.Equal(rr.Header().Get("X-Request-ID"), existingID)
}
