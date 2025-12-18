# Goberus

![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)

A minimal LDAP-backed service that exposes member lookup and provisioning workflows via `/v1/member`.

## Status
- [x] `GET /live` — liveness endpoint (always returns 200 OK with `{"status":"ok"}`)
- [x] `GET /ready` — readiness endpoint (returns 200 if LDAP is reachable, 503 otherwise)
- [x] `GET /v1/member?username=<value>` — resolves a user by UPN or sAMAccountName and returns normalized attributes via `server.UserClient` backed by `ldaps.Client` in production and fakes in tests.
- [x] `POST /v1/member` — sanitizes the JSON payload (trim + lowercase for `username`/`OrganizationalUnit`) with `handlers.SanitizeUser` before invoking `ldaps.Client.AddUser`.
- [ ] `DELETE /v1/member` — TODO: expose member removal once LDAP delete semantics and authorization are finalized.
- [ ] `PATCH /v1/member` — TODO: introduce attribute updates once LDAP modify flows are defined.

## Development & testing
See [docs/dev-setup.md](docs/dev-setup.md) for the quick-start instructions, environment variables, Docker guidance, troubleshooting tips, and the testing commands (`go test ./...`).

## Next steps
- Add API authentication and rate limiting
- Implement connection pooling/reconnect semantics
- Handle LDAPS password changes via `unicodePwd`
- Expand unit/integration coverage (e.g., Docker-compose with Samba AD)

## Project layout
```
.
├── cmd/
│   └── goberus/         # application entrypoint and process lifecycle
├── config/              # configuration loading and validation
├── internal/
│   └── httpserver/      # server composition, route wiring, JSON helpers
├── ldaps/               # LDAP client, models, helpers
├── middleware/          # HTTP middleware (RequestID, Recover, Logger)
├── server/              # HTTP handlers and server-facing types
├── handlers/            # auxiliary handler helpers used in tests/CLI
├── docs/                # developer and operational documentation
├── ADR/                 # Architecture Decision Records
├── Makefile
├── Dockerfile
└── README.md
```

## License
Goberus is open-source software distributed under the terms of the [GNU General Public License v3](LICENSE). See `LICENSE` for the full text and warranty disclaimer.
