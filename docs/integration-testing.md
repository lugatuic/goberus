# Integration Testing

This document describes how to run integration tests for Goberus against a live Samba Active Directory environment.

## Overview

Integration tests validate the full end-to-end flow of Goberus interacting with a real LDAP/AD backend. The tests use Docker Compose to spin up:
- **Samba AD Domain Controller** (nowsci/samba-domain) — provides a test Active Directory environment
- **Goberus service** — the application under test

## Prerequisites

- Docker and Docker Compose installed
- If you run Goberus directly on the host, ensure its chosen HTTP port is free (defaults to 8080).

## Secrets Management

**Local Development:** The `docker-compose.yml` file uses environment variables with safe defaults for local testing:
```bash
TEST_DOMAIN=TESTDOMAIN
TEST_DOMAIN_PASS=TestPass123!
TEST_BASE_DN=DC=testdomain,DC=local
TEST_BIND_DN=CN=Administrator,CN=Users,DC=testdomain,DC=local
TEST_BIND_PASSWORD=TestPass123!
```

These defaults are **only for local development** and are not production secrets.

**CI/CD (GitHub Actions):** In CI pipelines, these values **must** be set via GitHub Secrets and repository variables:
- `TEST_DOMAIN` → GitHub variable (non-sensitive)
- `TEST_DOMAIN_PASS` → GitHub secret
- `TEST_BASE_DN` → GitHub variable (non-sensitive)
- `TEST_BIND_DN` → GitHub variable (non-sensitive)
- `TEST_BIND_PASSWORD` → GitHub secret

See the CI Integration section below for workflow configuration.

## Running Integration Tests

### 1. Bring up the stack and run

```bash
# From repo root
docker compose up -d samba goberus
docker compose run --rm test-runner
docker compose down -v  # clean state
```

What this does:
- Starts Samba AD (LDAPS on host ports 389/636, self-signed cert).
- Starts Goberus on host port 8080; it talks to Samba with `LDAP_SKIP_VERIFY=true`.
- Runs the test runner inside the compose network (`go test ./tests/integration -v`).

Notes:
- Samba typically finishes provisioning within ~30s; `goberus` waits for its healthcheck, and the tests wait for `/readyz`.
- If you run tests outside Docker, set `INTEGRATION_BASE_URL` to your Goberus URL (e.g., `http://localhost:8080`).
- `docker compose down -v` resets AD state between runs.

## Test Coverage

The integration test suite (`tests/integration/integration_test.go`) validates:

### Health Endpoints
- `GET /livez` returns 200 with `{"status":"ok"}`
- `GET /readyz` returns 200 when LDAP is reachable with `{"status":"ready"}`

### Member Lookup (GET)
- `GET /v1/member?username=Administrator` resolves the default AD user
- Response is properly formatted `MemberInfo` JSON

### Member Creation (POST)
- `POST /v1/member` with valid payload creates a new user in AD
- `GET /v1/member` confirms the created user exists
- Username normalization (lowercase) is applied

### Request Correlation
- Custom `X-Request-ID` header is preserved in responses

### Error Handling
- Missing username parameter returns 400
- Invalid JSON payload returns 400
- Missing required fields returns 400

## Test Environment Details

### Samba AD Configuration
- **Domain:** Configured via `TEST_DOMAIN` env var (default: `TESTDOMAIN`)
- **Base DN:** Configured via `TEST_BASE_DN` (default: `DC=testdomain,DC=local`)
- **Admin User:** Configured via `TEST_BIND_DN` (default: `CN=Administrator,CN=Users,DC=testdomain,DC=local`)
- **Admin Password:** Configured via `TEST_DOMAIN_PASS` and `TEST_BIND_PASSWORD` (default: `TestPass123!` for local dev only)
- **Default OU for tests:** `CN=Users,DC=testdomain,DC=local`

### Goberus Configuration
Environment variables set in `docker-compose.yml`:
```env
LDAP_ADDR=samba:${TEST_LDAP_PORT:-636}
LDAP_HOST=samba
LDAP_PORT=${TEST_LDAP_PORT:-636}
LDAP_BASE_DN=${TEST_BASE_DN:-DC=testdomain,DC=local}
LDAP_BIND_DN=${TEST_BIND_DN:-CN=Administrator,CN=Users,DC=testdomain,DC=local}
LDAP_BIND_PASSWORD=${TEST_BIND_PASSWORD:-TestPass123!}
LDAP_USE_TLS=${TEST_LDAP_USE_TLS:-true}
LDAP_SKIP_VERIFY=${TEST_LDAP_SKIP_VERIFY:-true}
LDAP_CA_CERT=""
BIND_ADDR=:8080
```
`LDAP_SKIP_VERIFY` should be true locally because Samba uses a self-signed cert; you can supply a CA at `LDAP_CA_CERT` if desired.

## Troubleshooting

### Tests fail with "connection refused"
- Ensure `docker compose up` completed successfully
- Check `docker compose logs samba` for initialization errors
- Verify the Samba healthcheck passed: `docker compose ps`

### `/readyz` returns 503
- Ensure Samba is healthy.
- From host: `openssl s_client -connect localhost:636 -servername samba -brief` (expect self-signed warning but handshake should complete).
- From container: `docker compose exec -u root goberus sh -c "LDAPTLS_REQCERT=never ldapsearch -H ldaps://samba:636 -D 'CN=Administrator,CN=Users,DC=testdomain,DC=local' -w TestPass123! -b 'DC=testdomain,DC=local' -s base"`.
- If ldapsearch works, readyz should pass; otherwise inspect Samba logs.

### Tests timeout waiting for service
- Increase `maxRetries` or `retryInterval` in `integration_test.go`
- Check Goberus logs: `docker compose logs goberus`

### User creation fails
- Verify the bind credentials are correct
- Check LDAP logs: `docker compose logs samba`

### Clean state between test runs
Always run `docker compose down -v` to remove volumes and ensure a fresh AD instance.

## CI Integration

To run integration tests in CI (GitHub Actions), configure secrets and variables in your repository settings:

**GitHub Secrets (Settings → Secrets and variables → Actions → Secrets):**
- `TEST_DOMAIN_PASS` — Domain admin password
- `TEST_BIND_PASSWORD` — LDAP bind password (typically same as domain pass)

**GitHub Variables (Settings → Secrets and variables → Actions → Variables):**
- `TEST_DOMAIN` — Domain name (e.g., `TESTDOMAIN`)
- `TEST_BASE_DN` — Base DN (e.g., `DC=testdomain,DC=local`)
- `TEST_BIND_DN` — Bind DN (e.g., `CN=Administrator,CN=Users,DC=testdomain,DC=local`)

**Workflow example:**

```yaml
- name: Run integration tests
  env:
    TEST_DOMAIN: ${{ vars.TEST_DOMAIN }}
    TEST_DOMAIN_PASS: ${{ secrets.TEST_DOMAIN_PASS }}
    TEST_BASE_DN: ${{ vars.TEST_BASE_DN }}
    TEST_BIND_DN: ${{ vars.TEST_BIND_DN }}
    TEST_BIND_PASSWORD: ${{ secrets.TEST_BIND_PASSWORD }}
  run: |
    docker compose up -d
    sleep 30  # Wait for Samba initialization
    INTEGRATION_TESTS=true go test ./tests/integration -v
    docker compose down -v
```

See `.github/workflows/integration-tests.yml` for the full workflow.

## Extending Tests

To add new integration test cases:

1. Add test function to `tests/integration/integration_test.go`
2. Follow the pattern: call `waitForService(t)` at the start
3. Use `baseURL` constant for HTTP requests
4. Use `is.New(t)` for assertions
5. Clean up resources if creating test data

Example:
```go
func TestNewFeature(t *testing.T) {
    is := is.New(t)
    waitForService(t)
    
    resp, err := http.Get(baseURL + "/v1/newfeature")
    is.NoErr(err)
    defer resp.Body.Close()
    is.Equal(resp.StatusCode, http.StatusOK)
}
```
