# TODO / Issues

This file tracks planned features, improvements, and known issues for the Goberus project.

## Planned Features

### High Priority

- [ ] **Implement DELETE /v1/member endpoint**
  - Expose member removal functionality
  - Finalize LDAP delete semantics and authorization requirements
  - Add comprehensive test coverage

- [ ] **Implement PATCH /v1/member endpoint**
  - Introduce attribute update functionality
  - Define LDAP modify flows
  - Add comprehensive test coverage

### Medium Priority

- [ ] **Publish as GitHub Package**
  - Deferred until DELETE and PATCH endpoints are implemented
  - Set up GitHub Container Registry workflow
  - Create release automation
  - Document package installation and usage

- [ ] **Add API authentication and rate limiting**
  - Implement authentication middleware
  - Add rate limiting to prevent abuse
  - Document authentication requirements

- [ ] **Implement connection pooling/reconnect semantics**
  - Improve LDAP connection management
  - Add automatic reconnection logic
  - Add connection health monitoring

### Low Priority

- [ ] **Expand unit/integration coverage**
  - Add more comprehensive test scenarios
  - Improve test documentation

## Completed

- [x] Mat Ryer-style HTTP service architecture
- [x] Health endpoints (`/livez` and `/readyz`)
- [x] Request ID correlation middleware
- [x] GET /v1/member endpoint
- [x] POST /v1/member endpoint
- [x] LDAP client with LDAPS support
- [x] Docker multi-stage build
- [x] CI/CD workflows (tests, linting, CodeQL)
- [x] Handle LDAPS password changes via `unicodePwd`
- [x] Integration tests with Docker Compose and Samba AD
