# Project Structure Refactoring

## Overview
Reorganize project layout to improve clarity, maintainability, and discoverability of documentation, plans, and ADRs.

## Current Structure Issues
- **Scattered docs**: Configuration in `config/`, integration tests info in `docs/`, architecture decisions in `ADR/`
- **Hidden prompts**: Plans and agent prompts were nested in `.github/prompts`
- **Mixed concerns**: No clear separation between developer setup, operational docs, and design decisions
- **Growth friction**: As project adds LDAP features (PATCH, DELETE, Groups), docs will scatter further

## Proposed New Structure

```
goberus/
├── docs/                        # All documentation (consolidate here)
│   ├── architecture/            # Architecture decisions + design patterns
│   │   ├── ADR-001-middleware-design.md
│   │   ├── ADR-002-ldaps-operations.md  (NEW: refactoring rationale)
│   │   └── ...
│   ├── development/             # Developer-focused guides
│   │   ├── dev-setup.md
│   │   ├── integration-testing.md
│   │   ├── running-tests.md     (NEW: consolidate test instructions)
│   │   └── ...
│   ├── operations/              # Operational + deployment
│   │   ├── ldaps-cert-setup.md  (NEW: moved from Dockerfile comments)
│   │   ├── docker-compose-guide.md
│   │   └── ...
│   └── api/                     # API reference
│       ├── endpoints.md
│       ├── GET-member.md
│       ├── POST-member.md
│       ├── PATCH-member.md      (NEW: once implemented)
│       └── DELETE-member.md     (NEW: once implemented)
├── prompts/                     # Agent runbooks & implementation plans (moved from .github/prompts)
│   ├── Plan - LDAP Refactoring.md
│   ├── Plan - PATCH Method Support.md
│   ├── Plan - DELETE Method Support.md
│   ├── Plan - Project Structure Refactoring.md
│   └── Plan - IntegrationTestsAndCICD.md
├── http/                        # HTTP handlers & validation (NEW: consolidated from server/, handlers/)
│   ├── handlers.go              # HTTP handler functions (GET, POST, PATCH, DELETE /v1/member)
│   ├── handlers_test.go         # Handler unit tests
│   ├── validate.go              # Input validation helpers
│   └── validate_test.go         # NEW: Validation unit tests
├── middleware/                  # HTTP middleware (modular, pluggable)
│   ├── middleware.go            # Chain/composition utility
│   ├── recover.go               # Panic recovery middleware
│   ├── recover_test.go          # NEW: Recover tests
│   ├── requestid.go             # Request ID injection
│   ├── requestid_test.go        # NEW: RequestID tests
│   ├── logger.go                # REFACTOR: Logger as pluggable module (prep for OTel/Grafana)
│   └── logger_test.go           # NEW: Logger tests
├── ldaps/                       # LDAP operations (refactored by protocol)
│   ├── connection.go            # Bind operations
│   ├── search.go                # Search operations
│   ├── add.go                   # Add operations
│   ├── modify.go                # Modify operations (setUnicodePwd, ModifyUserAttributes)
│   ├── delete.go                # Delete operations (future)
│   ├── models.go
│   ├── dn.go
│   └── *_test.go
├── .github/
│   ├── workflows/
│   └── (no more /prompts here)
├── README.md
├── CHANGELOG.md
└── ... (rest of project)
```

## Implementation Steps

### Phase 1: Documentation Consolidation (this PR)
- [ ] Create `docs/` directory structure as above
- [ ] Move existing docs from `docs/` (keep as is for now)
- [ ] Create `docs/architecture/` and move ADR folder there (or symlink)
- [ ] Create `docs/development/` placeholder for future test docs
- [ ] Create `docs/api/` placeholder for future endpoint docs
- [ ] Move `prompts/` to root (already done in this commit)
- [ ] Update `README.md` to reference new docs structure
- [ ] Update `.gitignore` if needed

### Phase 1b: HTTP Handler Reorganization (same PR as Phase 1 docs, but separate commit)
- [ ] Create `http/` package to hold handlers and validation:
  - `http/handlers.go` — HTTP handler functions (GET, POST, PATCH, DELETE /v1/member)
  - `http/handlers_test.go` — Handler unit tests
  - `http/validate.go` — Input validation helpers
  - `http/validate_test.go` — **NEW**: Validation tests
- [ ] Refactor `middleware/` into modular, pluggable middleware:
  - `middleware/middleware.go` — Chain/composition utility (keep as middleware glue)
  - `middleware/recover.go` — Panic recovery middleware
  - `middleware/recover_test.go` — **NEW**: Recover tests
  - `middleware/requestid.go` — Request ID injection (keep existing, add tests)
  - `middleware/requestid_test.go` — **NEW**: RequestID tests
  - `middleware/logger.go` — **REFACTOR**: Extract Logger into own module (preparing for future OTel/Grafana integration)
  - `middleware/logger_test.go` — **NEW**: Logger tests (enables testing OTel hooks)
- [ ] Delete now-empty `server/` and `handlers/` directories
- [ ] Update imports in `internal/httpserver/` and `cmd/goberus/main.go`
- [ ] Verify `go test ./...` passes with full coverage

**Rationale**: 
- Handlers (http/) and validation (http/) are colocated but separate from infrastructure concerns
- Middleware stays in middleware/ package, but each middleware is independently pluggable
- Logger isolation enables future extensions (OTel exporter, Grafana integration, custom hooks)
- Individual middleware_test.go files enable comprehensive testing and easier mocking
- Middleware composition in middleware.go remains lightweight coordination layer

### Phase 2: LDAP Refactoring (separate PR)
- [ ] Create `ldaps/connection.go` — Bind/connection operations
- [ ] Create `ldaps/search.go` — Move GetMemberInfo, Ping (verify)
- [ ] Rename `ldaps/add_user.go` → `ldaps/add.go`
- [ ] Create `ldaps/modify.go` — Move setUnicodePwd, enableAccount; add ModifyUserAttributes
- [ ] Create `ldaps/delete.go` — Placeholder for DeleteUser
- [ ] Update imports across http/ and internal/ packages
- [ ] Run tests to verify no breakage

### Phase 3: PATCH Support (separate PR)
- [ ] Add `HandlePatchMember` in `server/handlers.go`
- [ ] Add unit tests for PATCH handler
- [ ] Add integration tests for PATCH flows
- [ ] Update `docs/api/PATCH-member.md`

### Phase 4: DELETE Support (separate PR)
- [ ] Add `HandleDeleteMember` in `server/handlers.go`
- [ ] Add unit tests for DELETE handler
- [ ] Add integration tests for DELETE flows
- [ ] Update `docs/api/DELETE-member.md`

## Rationale

### Why consolidate docs?
- **Single source of truth**: All documentation in one place (docs/)
- **Better discoverability**: New contributors find everything in docs/, not scattered
- **Easier navigation**: Architecture decisions, dev guides, and API reference coexist
- **Growth-ready**: PATCH, DELETE, group management, and future features have designated spots

### Why move prompts to root?
- **Visibility**: Plans and runbooks are at project root, easy to find
- **Agent-friendly**: Simpler path references in prompts (../prompts/Plan-*.md vs ../.github/prompts/...)
- **Keeps .github/ clean**: GitHub workflows config separate from project content

### Why refactor ldaps/ by protocol?
- **Maintainability**: All modify operations (setUnicodePwd, enableAccount, ModifyUserAttributes) in one file
- **Scalability**: Adding DELETE, future operations fits naturally
- **Clarity**: Each file represents one LDAP operation type (Search, Add, Modify, Delete)

## Files Changed
| Path | Action | Note |
|------|--------|------|
| `.github/prompts/` | Delete | Moved to `prompts/` |
| `prompts/` | Create | Root-level plans and agent runbooks |
| `docs/` | Reorganize | Add subdirs for architecture, development, operations, api |
| `ADR/` | Relocate | Move or symlink to `docs/architecture/ADR/` |
| `ldaps/*.go` | Refactor | Split into connection, search, add, modify, delete (Phase 2) |

## Validation & Testing
- [ ] No files lost or corrupted
- [ ] All imports updated
- [ ] `go test ./...` passes
- [ ] `INTEGRATION_TESTS=true go test ./tests/integration -v` passes (if Samba running)
- [ ] Links in docs are updated if needed
- [ ] Verify agents can still find prompts at new location

## Timeline
- **Phase 1 (Docs)**: This PR
- **Phase 2 (LDAP refactor)**: Follow-up PR
- **Phase 3 (PATCH)**: Follow-up PR
- **Phase 4 (DELETE)**: Follow-up PR

Estimated: 2-3 weeks to complete all phases.
