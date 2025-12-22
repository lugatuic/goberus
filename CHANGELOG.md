# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.3]

### Added
- Mat Ryer-style HTTP service architecture with graceful shutdown
- Health endpoints: `/livez` and `/readyz` (LDAP connectivity check)
- Request ID correlation middleware for distributed tracing
- Versioning support via ldflags injection in Docker builds
- Architecture Decision Record for middleware design pattern

### Changed
- Restructured entrypoint: moved to `cmd/goberus/main.go`
- Updated project layout with clear package boundaries
- Enhanced Makefile with version and release management

## [0.0.2] - 2025-12-17

### Added
- `POST /v1/member` endpoint for user provisioning
- Unicode password encoding support for Active Directory
- Comprehensive test coverage for handlers and LDAP operations
- CI/CD workflows: Go tests, linting, and CodeQL security analysis

### Changed
- Enhanced LDAP client with organizational unit (OU) support
- Integrated Zap structured logging

## [0.0.1] - 2025-11-18

### Added
- `GET /v1/member` endpoint for member lookup
- LDAP client with connection pooling and LDAPS support
- Docker multi-stage build with Alpine runtime
- Basic project structure and configuration
