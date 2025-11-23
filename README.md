# Goberus

A minimal LDAP-backed service that exposes member lookup and provisioning workflows via `/v1/member`.

## Status
- [x] `GET /v1/member?username=<value>` — resolves a user by UPN or sAMAccountName and returns normalized attributes via `server.UserClient` backed by `ldaps.Client` in production and fakes in tests.
- [x] `POST /v1/member` — sanitizes the JSON payload (trim + lowercase for `username`/`OrganizationalUnit`) with `handlers.SanitizeUser` before invoking `ldaps.Client.AddUser`.
- [ ] `DELETE /v1/member` — TODO: expose member removal once LDAP delete semantics and authorization are finalized.
- [ ] `PATCH /v1/member` — TODO: introduce attribute updates once LDAP modify flows are defined.
- `tests/server/handlers_test.go` covers the handler flows (including an integration-style `httptest.NewServer` canary), and `handlers/users_test.go` focuses on `SanitizeUser`.

## Development & testing
See [docs/dev-setup.md](docs/dev-setup.md) for the quick-start instructions, environment variables, Docker guidance, troubleshooting tips, and the testing commands (`go test ./tests/server -run TestHandleGetMember`, `go test ./tests/server -run TestHandleCreateMember`, `go test ./tests/server -run TestSanitizeUserIntegration`, `go test ./...`).

## Next steps
- Add API authentication and rate limiting
- Implement connection pooling/reconnect semantics
- Handle LDAPS password changes via `unicodePwd`
- Expand unit/integration coverage (e.g., Docker-compose with Samba AD)

## License
Goberus is open-source software distributed under the terms of the [GNU General Public License v3](LICENSE). See `LICENSE` for the full text and warranty disclaimer.
