# Agent Workflow Notes

**Purpose**: Keep our agents consistent when working on this repo. Always verify changes with Dockerized integration tests before committing.

## Ground rules
- Scan existing docs and directories before making changes.
- Do not create stub docs or empty directories.
- Do not duplicate content across locations; move files and update links.
- Make the smallest change that satisfies the request.
- Keep documentation organized under `docs/development/` and `docs/architecture/ADR/`.
- Avoid adding new top-level directories without explicit approval.
- Use the changelog for historical record; don't add extra "notes" files for work summaries.
- Verify links after moves.

## Existing documentation (scan first)
- Development setup and environment variables: [dev-setup.md](dev-setup.md)
- Integration testing workflow and CI notes: [integration-testing.md](integration-testing.md)
- ADR template and decision format: [ADR/0000_template.md](../architecture/ADR/0000_template.md)
- Middleware structure decision (Mat Ryer pattern): [ADR/2025-12-18_middleware_design_pattern.md](../architecture/ADR/2025-12-18_middleware_design_pattern.md)
- Documentation consolidation decision: [ADR/2025-12-22_consolidate_documentation_structure.md](../architecture/ADR/2025-12-22_consolidate_documentation_structure.md)

## Standard Workflow
1. **Clean slate (optional but recommended)**
   - `docker compose down -v || true`
   - `docker system prune -f` (only if you need to reclaim space)

2. **Rebuild app image and start services**
   - `docker compose build goberus`
   - `docker compose up -d --build samba goberus`

3. **Wait for health**
   - Verify both services are healthy:
     - `docker compose ps`
     - Samba should show healthy on 389/636; goberus healthy on 8080.

4. **Run golangci-lint (MANDATORY)**
   - `golangci-lint run ./...` (ensure it’s installed, e.g., `brew install golangci-lint`).

5. **Run integration tests (MANDATORY BEFORE ANY COMMIT)**
   - `docker compose run --rm test-runner`
   - All tests must pass before staging/committing.

6. **If tests fail**
   - Inspect logs: `docker compose logs samba --tail=200` and `docker compose logs goberus --tail=200`
   - Fix issues, rebuild (`docker compose up -d --build samba goberus`), and rerun tests.

7. **Commit only after green tests**
   - `git status` should reflect intended changes only.
   - Stage/commit after step 4 passes.

8. **Cleanup when done**
   - `docker compose down -v` to tear down services and volumes.

## Notes
- Compose defaults: Samba uses ports 389/636, goberus on 8080. Avoid host port conflicts.
- Go toolchain: Docker builder uses Go 1.23 to match `go.mod`.
- Prefer minimal changes; don’t modify DNS forwarder or healthchecks unless necessary.- When releasing, bump `version.txt` and update `CHANGELOG.md`. Use `make bump-version VERSION=x.y.z` or edit directly.- Refer to the integration test workflow if necessary: [Integration Testing](integration-testing.md)
