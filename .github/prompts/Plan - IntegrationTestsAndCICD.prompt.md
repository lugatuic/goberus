# Plan: Integration tests + CI/CD pipeline + versioning

**TL;DR:** Establish semantic versioning and git tagging; create a docker-compose suite for Samba integration tests; add a GitHub Actions workflow to build, tag, and publish container images to GitHub Container Registry (ghcr.io); retroactively document releases via CHANGELOG.

## Steps

1. **Define versioning strategy**: Adopt semantic versioning starting from v0.0.1 (GET endpoint); inject version into binary via `ldflags`; establish git tag naming. Version increments: v0.0.1 (GET), v0.0.2 (POST), v0.1.0 (first minor feature), etc.

2. **Create integration test infrastructure**: Write `docker-compose.yml` with Samba AD container, fixture setup, and test orchestration; add `tests/integration/` directory with Go test suite.

3. **Implement version endpoint**: Add `/version` as a separate endpoint (distinct from health checks) to expose build version, commit hash, and build time.

4. **Add GitHub Actions workflow for CI/CD**: Create `build-publish.yml` that builds multi-platform images, tags with version/`latest`, and pushes to `ghcr.io/lugatuic/goberus`.

5. **Create retroactive CHANGELOG**: Document past commits/features since project inception; establish format for future releases.

6. **Update Makefile & docs**: Add `make version`, `make release` targets; document the release process in `docs/`.

## Decisions on Further Considerations

1. **Git tagging & release triggers**: Releases automated on merge to `main`
   - **Rationale:** If integration tests aren't passing, code shouldn't merge to `main` anyway. Automates the release process and ensures every merge to `main` is production-ready.
   - **Implementation:** `build-publish.yml` triggers on `push` to `main` and publishes immediately.

2. **Multi-platform builds**: Yes, publish for `linux/amd64` and `linux/arm64` via `docker/build-push-action`
   - **Benefit:** Supports x86 and Apple Silicon (ARM) developers; future-proof for AWS Graviton and other ARM deployments.
   - **Implementation:** Use `docker/setup-buildx-action` and set `platforms: linux/amd64,linux/arm64` in build-push step.

3. **Samba test environment scope**: Minimal fixture with coverage for implemented functionality
   - **Scope:** Single test user (e.g., `testuser`), OU (`ou=testing`), basic group membership
   - **Functionality covered:** 
     - GET `/v1/member?username=testuser` resolves successfully
     - POST `/v1/member` with valid payload creates new user
     - Health check `/ready` returns 200 when Samba is reachable
   - **Future:** Expand as PATCH/DELETE and group management features are added.

4. **Changelog format**: Keep-a-Changelog (https://keepachangelog.com/)
   - **Structure:** Sections for `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security` per release
   - **Versioning link:** Semantic versioning (v0.0.1, v0.0.2, v0.1.0, etc.)
   - **Example format:**
     ```markdown
     ## [0.0.2] - 2025-12-18
     ### Added
     - POST `/v1/member` endpoint for user provisioning with sanitization
     - Integration tests against Samba fixture
     - `/version` endpoint exposing build metadata
     
     ### Changed
     - Restructured HTTP server following Mat Ryer patterns
     
     ## [0.0.1] - 2025-12-17
     ### Added
     - GET `/v1/member` endpoint for user lookup
     - Health checks: `/live` and `/ready`
     - Request ID correlation middleware
     ```

## Implementation Timeline

### Phase 1: Foundation (versioning + Makefile)
- Add `VERSION` file (v0.0.1)
- Modify `Dockerfile` to inject version via `ldflags`
- Update `Makefile` with `version` and `release` targets
- Create `CHANGELOG.md` (retroactive entries for v0.0.1: GET endpoint, v0.0.2: POST endpoint)

### Phase 2: Integration tests
- Create `docker-compose.yml` (Samba AD fixture)
- Add `tests/integration/` test suite
- Document test setup in `docs/integration-testing.md`
- Add integration test step to CI workflow (or separate workflow)

### Phase 3: CI/CD pipeline
- Create `.github/workflows/build-publish.yml`
- Configure multi-platform builds
- Add GitHub Container Registry (ghcr.io) push
- Test release process on feature branch

### Phase 4: Documentation & refinement
- Update `docs/dev-setup.md` with release process
- Add version endpoint to health checks (if desired)
- Document versioning strategy in new `ADR/2025-12-18_versioning_and_release.md`

## Key Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `VERSION` | Create | Semantic version source of truth |
| `CHANGELOG.md` | Create | Release history and features |
| `docker-compose.yml` | Create | Samba AD test environment |
| `tests/integration/` | Create | Integration test suite |
| `Dockerfile` | Modify | Inject version via `ldflags` |
| `Makefile` | Modify | Add version/release targets |
| `.github/workflows/build-publish.yml` | Create | Multi-platform build and publish |
| `docs/integration-testing.md` | Create | Integration test documentation |
| `docs/release-process.md` | Create | Release workflow documentation |

## Success Criteria

- [ ] Binary includes version info (visible in `/version` endpoint or build output)
- [ ] Integration tests run against Samba SUT in docker-compose
- [ ] CI/CD workflow builds and publishes multi-platform images to ghcr.io on release
- [ ] CHANGELOG reflects past features and establishes format for future releases
- [ ] All changes backward-compatible; no breaking changes to existing API

## Decisions on Open Questions

### 1. Version endpoint: Integrated vs. separate

**Decision:** Add `/version` as a **separate endpoint** (not part of existing health checks).

**Rationale & Examples:**

- **Option A (Integrated — rejected):** Include version in `/live` or `/ready` response:
  ```json
  GET /live
  {
    "status": "ok",
    "version": "0.0.2",
    "commit": "abc1234"
  }
  ```
  *Problem:* Health checks should be minimal and fast; version info is informational, not critical to liveness.

- **Option B (Separate — chosen):** Dedicated `/version` endpoint:
  ```json
  GET /version
  {
    "version": "0.0.2",
    "commit": "abc1234",
    "buildTime": "2025-12-18T14:30:00Z"
  }
  ```
  *Benefit:* Clean separation of concerns; tools like Kubernetes can poll `/live` for health without receiving version metadata; ops can query `/version` independently.

### 2. Samba image pinning

**Decision:** Pin Samba image to `latest` in docker-compose (no specific version tag).

**Rationale:** Keeps test environment up-to-date with Samba security patches; integration tests are not part of production release artifact, so flexibility is acceptable. Revisit if flakiness emerges.

### 3. Branch protection rules

**Decision:** Require version bump + CHANGELOG entry to merge to `main`.

**Implementation:**
- Add branch protection rule on `main` requiring:
  - At least one approval
  - All status checks pass (lint, test, build)
  - Require dismissal of stale reviews
- Pre-merge validation (via CI): Check that `VERSION` file changed and `CHANGELOG.md` updated
- Document this in `docs/release-process.md`

**Rationale:** Enforces disciplined releases; prevents accidental semver violations; maintains clear release history.
