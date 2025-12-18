# ADR: Align HTTP service architecture with Mat Ryer's guidance

Status: Accepted
Date: 2025-12-18

## Context
The current service uses `net/http` with custom middleware and a root-level `main.go`. While functional, it lacks:
- Graceful shutdown and signal handling.
- Health/readiness endpoints for orchestration.
- Hardened server timeouts.
- Request correlation (request IDs).
- Clear separation of entrypoint and server composition.

Mat Ryer's article, [How I write HTTP services in Go after 13 years](https://grafana.com/blog/how-i-write-http-services-in-go-after-13-years/), advocates a pragmatic, stdlib-first approach with explicit middleware, clean separation of concerns, graceful shutdown, health endpoints, and sensible timeouts.

## Decision
Adopt Mat Ryer–style architecture:
- Introduce a dedicated entrypoint at `cmd/goberus/main.go` that configures an `http.Server` with `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`, starts the server in a goroutine, and performs graceful shutdown on SIGINT/SIGTERM.
- Add `internal/httpserver` package to compose dependencies, wire routes, and enforce JSON responses, avoiding leakage of internal error details.
- Add health endpoints: `/live` (always 200) and `/ready` (checks LDAP via a new `ldaps.Client.Ping(ctx)`).
- Add `middleware.RequestID` and order middleware: `Recover` (outermost) → `RequestID` → `Logger`.
- Remove the root-level `main.go` to consolidate the entrypoint.

## Consequences
- Operational reliability improves (graceful shutdown, clear health/readiness signals).
- Security and robustness improve (timeouts; no internal error strings in responses).
- Observability improves (stable request correlation via `X-Request-ID`).
- Clearer structure for future features (metrics/tracing, CLI flags), consistent with Mat Ryer's recommendations.

## Alternatives considered
- Keep single-file root main: simpler but lacks lifecycle and health primitives.
- Adopt a heavy HTTP framework: unnecessary; stdlib-first approach is sufficient and preferred.

## Follow-ups
- Consider migrating non-public packages under `internal/` (e.g., `internal/ldaps`, `internal/server`, `internal/config`) to mark boundaries.
- Add CLI flags (with env overrides) for configuration discovery.
- Add basic metrics (e.g., Prometheus) for requests, latencies, and error rates.
- Expand JSON error handling with well-defined error types and status mapping.
