# Agent Workflow Notes

**Purpose**: Keep our agents consistent when working on this repo. Always verify changes with Dockerized integration tests before committing.

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

4. **Run integration tests (MANDATORY BEFORE ANY COMMIT)**
   - `docker compose run --rm test-runner`
   - All tests must pass before staging/committing.

5. **If tests fail**
   - Inspect logs: `docker compose logs samba --tail=200` and `docker compose logs goberus --tail=200`
   - Fix issues, rebuild (`docker compose up -d --build samba goberus`), and rerun tests.

6. **Commit only after green tests**
   - `git status` should reflect intended changes only.
   - Stage/commit after step 4 passes.

7. **Cleanup when done**
   - `docker compose down -v` to tear down services and volumes.

## Notes
- Compose defaults: Samba uses ports 389/636, goberus on 8080. Avoid host port conflicts.
- Go toolchain: Docker builder uses Go 1.23 to match `go.mod`.
- Prefer minimal changes; don’t modify DNS forwarder or healthchecks unless necessary.
- Refer to the integration test workflow if necessary: [Integration Testing](../docs/integration-testing.md)