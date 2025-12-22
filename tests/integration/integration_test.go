package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/matryer/is"

	"github.com/lugatuic/goberus/ldaps"
)

const (
	defaultBaseURL = "http://localhost:8080"
	maxRetries     = 30
	retryInterval  = 2 * time.Second
)

func uniqueUsername(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func baseURL() string {
	if v := os.Getenv("INTEGRATION_BASE_URL"); v != "" {
		return v
	}
	return defaultBaseURL
}

func closeBody(t *testing.T, closer io.Closer) {
	t.Helper()
	if err := closer.Close(); err != nil {
		t.Errorf("close response body: %v", err)
	}
}

type testLogger struct{ t *testing.T }

func (l testLogger) Printf(format string, args ...any) {
	l.t.Helper()
	l.t.Logf(format, args...)
}

// waitForService polls the service until it's ready or times out
func waitForService(t *testing.T) {
	t.Helper()
	is := is.New(t)

	client := retryablehttp.NewClient()
	client.RetryMax = maxRetries
	client.RetryWaitMin = retryInterval
	client.RetryWaitMax = retryInterval
	client.Logger = testLogger{t: t}
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return true, nil
		}
		if resp == nil {
			return true, nil
		}
		if resp.StatusCode == http.StatusOK {
			return false, nil
		}
		// Drain/close before retry to avoid leaks
		closeBody(t, resp.Body)
		return true, nil
	}

	req, err := retryablehttp.NewRequest(http.MethodGet, baseURL()+"/readyz", nil)
	is.NoErr(err)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Service did not become ready after %d attempts: %v", client.RetryMax+1, err)
	}
	if resp != nil {
		defer closeBody(t, resp.Body)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Service not ready after %d attempts: status=%d body=%s", client.RetryMax+1, resp.StatusCode, string(body))
	}
	t.Logf("Service ready (status=%d)", resp.StatusCode)
}

func TestMain(m *testing.M) {
	// Skip integration tests if not running in docker-compose environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		fmt.Println("Skipping integration tests (set INTEGRATION_TESTS=true to run)")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

func TestHealthEndpoints(t *testing.T) {
	is := is.New(t)
	waitForService(t)

	t.Run("/livez returns 200", func(t *testing.T) {
		resp, err := http.Get(baseURL() + "/livez")
		is.NoErr(err)
		defer closeBody(t, resp.Body)
		is.Equal(resp.StatusCode, http.StatusOK)

		var body map[string]string
		is.NoErr(json.NewDecoder(resp.Body).Decode(&body))
		is.Equal(body["status"], "ok")
	})

	t.Run("/readyz returns 200 when LDAP is reachable", func(t *testing.T) {
		resp, err := http.Get(baseURL() + "/readyz")
		is.NoErr(err)
		defer closeBody(t, resp.Body)
		is.Equal(resp.StatusCode, http.StatusOK)

		var body map[string]string
		is.NoErr(json.NewDecoder(resp.Body).Decode(&body))
		is.Equal(body["status"], "ready")
	})
}

func TestGetMemberIntegration(t *testing.T) {
	is := is.New(t)
	waitForService(t)

	// Query for the default Administrator user (should exist in Samba)
	resp, err := http.Get(baseURL() + "/v1/member?username=Administrator")
	is.NoErr(err)
	defer closeBody(t, resp.Body)

	is.Equal(resp.StatusCode, http.StatusOK)

	var member ldaps.MemberInfo
	is.NoErr(json.NewDecoder(resp.Body).Decode(&member))
	is.Equal(member.Username, "administrator") // Normalized to lowercase
}

func TestCreateMemberIntegration(t *testing.T) {
	is := is.New(t)
	waitForService(t)

	username := uniqueUsername("integrationuser")
	password := "IntegTest123!"

	// Create a new user
	payload := map[string]string{
		"username": username,
		"password": password,
	}
	body, err := json.Marshal(payload)
	is.NoErr(err)

	resp, err := http.Post(baseURL()+"/v1/member", "application/json", bytes.NewReader(body))
	is.NoErr(err)
	defer closeBody(t, resp.Body)

	is.Equal(resp.StatusCode, http.StatusCreated)

	var result map[string]string
	is.NoErr(json.NewDecoder(resp.Body).Decode(&result))
	is.Equal(result["status"], "created")

	// Verify the user was created by querying it
	t.Run("verify created user", func(t *testing.T) {
		resp, err := http.Get(baseURL() + "/v1/member?username=" + username)
		is.NoErr(err)
		defer closeBody(t, resp.Body)

		is.Equal(resp.StatusCode, http.StatusOK)

		var member ldaps.MemberInfo
		is.NoErr(json.NewDecoder(resp.Body).Decode(&member))
		is.Equal(member.Username, username)
	})
}

func TestCreateMemberUnicodePassword(t *testing.T) {
	is := is.New(t)
	waitForService(t)

	username := uniqueUsername("unicodeuser")
	password := "AÜnicodePwd123!"

	payload := map[string]string{
		"username": username,
		"password": password,
	}
	body, err := json.Marshal(payload)
	is.NoErr(err)

	resp, err := http.Post(baseURL()+"/v1/member", "application/json", bytes.NewReader(body))
	is.NoErr(err)
	defer closeBody(t, resp.Body)

	is.Equal(resp.StatusCode, http.StatusCreated)

	var result map[string]string
	is.NoErr(json.NewDecoder(resp.Body).Decode(&result))
	is.Equal(result["status"], "created")

	t.Run("verify created unicode user", func(t *testing.T) {
		resp, err := http.Get(baseURL() + "/v1/member?username=" + username)
		is.NoErr(err)
		defer closeBody(t, resp.Body)

		is.Equal(resp.StatusCode, http.StatusOK)

		var member ldaps.MemberInfo
		is.NoErr(json.NewDecoder(resp.Body).Decode(&member))
		is.Equal(member.Username, username)
	})
}

func TestRequestIDCorrelation(t *testing.T) {
	is := is.New(t)
	waitForService(t)

	req, err := http.NewRequest(http.MethodGet, baseURL()+"/livez", nil)
	is.NoErr(err)

	// Send request with custom X-Request-ID
	customID := "test-correlation-123"
	req.Header.Set("X-Request-ID", customID)

	client := &http.Client{}
	resp, err := client.Do(req)
	is.NoErr(err)
	defer closeBody(t, resp.Body)

	is.Equal(resp.StatusCode, http.StatusOK)
	is.Equal(resp.Header.Get("X-Request-ID"), customID)
}

func TestInvalidInputHandling(t *testing.T) {
	is := is.New(t)
	waitForService(t)

	t.Run("missing username parameter", func(t *testing.T) {
		resp, err := http.Get(baseURL() + "/v1/member")
		is.NoErr(err)
		defer closeBody(t, resp.Body)

		is.Equal(resp.StatusCode, http.StatusBadRequest)
		body, _ := io.ReadAll(resp.Body)
		is.True(bytes.Contains(body, []byte("missing username parameter")))
	})

	t.Run("invalid json payload", func(t *testing.T) {
		resp, err := http.Post(baseURL()+"/v1/member", "application/json", bytes.NewReader([]byte("{invalid")))
		is.NoErr(err)
		defer closeBody(t, resp.Body)

		is.Equal(resp.StatusCode, http.StatusBadRequest)
		body, _ := io.ReadAll(resp.Body)
		is.True(bytes.Contains(body, []byte("invalid json")))
	})

	t.Run("missing required fields", func(t *testing.T) {
		payload := map[string]string{"username": "a"}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(baseURL()+"/v1/member", "application/json", bytes.NewReader(body))
		is.NoErr(err)
		defer closeBody(t, resp.Body)

		is.Equal(resp.StatusCode, http.StatusBadRequest)
	})
}
