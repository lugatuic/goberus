# Middleware design pattern

## Status
Accepted (2025-12-18)

## Context
The service currently uses `net/http` with custom middleware and a root-level `main.go` but lacks graceful shutdown, health/readiness endpoints, hardened timeouts, request correlation, and a clean separation between entrypoint and server composition. Mat Ryer's article, [How I write HTTP services in Go after 13 years](https://grafana.com/blog/how-i-write-http-services-in-go-after-13-years/), advocates a pragmatic stdlib-first approach with explicit middleware, lifecycle management, and sensible defaults.

## Decision
Adopt the Mat Ryer–style HTTP service structure:
- Provide a dedicated entrypoint at `cmd/goberus/main.go` that configures `http.Server` timeouts, starts the server in a goroutine, and performs graceful shutdown on SIGINT/SIGTERM.
- Add `internal/httpserver` to compose dependencies, wire routes, and enforce JSON responses without leaking internal errors.
- Expose health endpoints: `/livez` (always 200) and `/readyz` (checks LDAP via `ldaps.Client.Ping(ctx)`).
- Apply middleware ordering: `Recover` (outermost) → `RequestID` → `Logger` to ensure correlation IDs and stable logging.
- Remove the legacy root `main.go` in favor of the structured entrypoint.

## Consequences
- Improves operational reliability via graceful shutdown and clear health/readiness signals.
- Hardens security and robustness with timeouts and controlled error responses.
- Enhances observability with consistent request correlation (`X-Request-ID`).
- Establishes a clearer foundation for future metrics, tracing, and configuration work while keeping a stdlib-first footprint.
