# Goberus

![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)

A minimal LDAP-backed service that exposes member lookup and provisioning workflows via `/v1/member`.

**Architecture**: Following [Mat Ryer's HTTP service patterns](https://grafana.com/blog/how-i-write-http-services-in-go-after-13-years/), the service uses a dedicated entrypoint at `cmd/goberus/main.go` with graceful shutdown, hardened timeouts, health endpoints, and request correlation via `X-Request-ID` headers. See [ADR.md](ADR.md) for architectural decisions.

## Status
- [x] `GET /live` — liveness endpoint (always returns 200 OK with `{"status":"ok"}`)
- [x] `GET /ready` — readiness endpoint (returns 200 if LDAP is reachable, 503 otherwise)
- [x] `GET /v1/member?username=<value>` — resolves a user by UPN or sAMAccountName and returns normalized attributes via `server.UserClient` backed by `ldaps.Client` in production and fakes in tests.
- [x] `POST /v1/member` — sanitizes the JSON payload (trim + lowercase for `username`/`OrganizationalUnit`) with `handlers.SanitizeUser` before invoking `ldaps.Client.AddUser`.
- [ ] `DELETE /v1/member` — TODO: expose member removal once LDAP delete semantics and authorization are finalized.
- [ ] `PATCH /v1/member` — TODO: introduce attribute updates once LDAP modify flows are defined.

## Development & testing
See [docs/dev-setup.md](docs/dev-setup.md) for the quick-start instructions, environment variables, Docker guidance, troubleshooting tips, and the testing commands (`go test ./tests/server -run TestHandleGetMember`, `go test ./tests/server -run TestHandleCreateMember`, `go test ./tests/server -run TestSanitizeUserIntegration`, `go test ./...`).

## Next steps
- Add API authentication and rate limiting
- Implement connection pooling/reconnect semantics
- Handle LDAPS password changes via `unicodePwd`
- Expand unit/integration coverage (e.g., Docker-compose with Samba AD)

## License
Goberus is open-source software distributed under the terms of the [GNU General Public License v3](LICENSE). See `LICENSE` for the full text and warranty disclaimer.
