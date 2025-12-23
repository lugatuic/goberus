# PATCH /v1/member — Attribute-level modify support

## Overview
Add `PATCH /v1/member` endpoint for attribute-level user modifications in Active Directory. Re-uses existing `ldaps` unicodePwd helpers and test patterns.

## Existing code to reuse
- `ldaps/unicode_pwd_test.go`: mockModifier pattern for asserting ModifyRequest structure
- `ldaps/unicode_pwd.go`: encodeUnicodePwd, setUnicodePwd, enableAccount helpers
- `server/handlers_test.go`: fakeClient pattern for handler testing
- `tests/integration/integration_test.go`: Samba AD test harness and INTEGRATION_TESTS gate

## Implementation tasks

### 1. ldaps-level modify support (`ldaps/modify.go` + tests)
- [ ] Implement `Client.ModifyUserAttributes(ctx context.Context, username string, ops []ModifyOperation) error`
  - Accepts a list of operations (add/replace/delete) for specified attributes
  - Reuses mockModifier pattern for unit testing
  - Unit tests: TestModifyUserAttributes_BuildsModifyRequest, TestModifyUserAttributes_PropagatesErrors

### 2. HTTP PATCH handler (`server/handlers.go` or `server/patch.go`)
- [ ] Implement `HandlePatchMember(client UserClient, w http.ResponseWriter, r *http.Request) error`
  - Parse JSON: `{"username":"alice","changes":[{"op":"replace","attr":"displayName","values":["Alice Q"]}]}`
  - Validate username, ops, attribute whitelist
  - Call `client.ModifyUserAttributes()`
  - Return HTTP 200 with updated user object (via GET) on success; 400 on validation error, 500 on backend error
- [ ] Unit tests: invalid JSON → 400, invalid op → 400, success → 200 + user object, backend error → 500

### 3. Router wiring
- [ ] Update `server/handlers.go` UserClient interface: add `ModifyUserAttributes()` method
- [ ] Update `internal/httpserver` UserClient interface: add `ModifyUserAttributes()` method
- [ ] Wire `HandlePatchMember` into HTTP router: add `case http.MethodPatch` in handler switch

### 4. Integration tests (`tests/integration/patch_test.go`)
- [ ] TestPatchMemberIntegration_ReplaceAttribute: create user → PATCH displayName → GET verify
- [ ] TestPatchMemberIntegration_AddAttribute: create user → PATCH add telephoneNumber → GET verify
- [ ] Gate with `INTEGRATION_TESTS=true`

### 5. Documentation
- [ ] Update `README.md`: document PATCH endpoint and request/response format
- [ ] Update `docs/integration-testing.md`: instructions to run integration tests locally

## Attribute whitelist
Allowed for patching: displayName, givenName, sn, mail, telephoneNumber, description, physicalDeliveryOfficeName, streetAddress, l, st, postalCode, c, co, department, title, company, manager, extensionAttributes

Rejected: objectClass, memberOf, userAccountControl, unicodePwd (use separate password change flow)

## Testing checklist
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (all unit tests)
- [ ] `INTEGRATION_TESTS=true go test ./tests/integration -v` passes with Samba AD running

## Files to create/modify
| File | Action | Notes |
|------|--------|-------|
| `ldaps/modify.go` | Create | ModifyUserAttributes + ModifyOperation type |
| `ldaps/modify_test.go` | Create | Unit tests using mockModifier |
| `server/handlers.go` or `server/patch.go` | Create/Modify | HandlePatchMember + update UserClient interface |
| `internal/httpserver` | Modify | Update UserClient interface + router case |
| `tests/integration/patch_test.go` | Create | Integration tests |
| `README.md` | Modify | Document PATCH endpoint |
| `docs/integration-testing.md` | Modify | Add local test instructions |

## Security notes
- Enforce LDAPS/TLS for all PATCH operations
- Never log attribute values that may contain sensitive data
- Whitelist approach: only allow explicitly-approved attributes

## Plan: Patch Support (attribute-level modify)

TL;DR — Add attribute-level PATCH support by re-using the existing `ldaps` unicodePwd helpers and test mocks, implementing a small ldaps modify helper, a `PATCH /v1/member` HTTP handler, unit tests (mockModifier/fakeClient pattern), and integration tests re-using the current Samba AD harness. The repo already contains password helpers, modify-test patterns, and integration scaffolding; gaps include an exported ldaps modify API, a handler and router case for PATCH, and unit+integration tests for patch flows.

### Steps
1. Add ldaps modify helper: implement `Client.ModifyUserAttributes` in `ldaps/modify.go` (accept DN or username + []ModifyOperation).
2. Add unit tests for ldaps modify logic: `ldaps/modify_test.go` (use `mockModifier` pattern).
3. Add server handler for PATCH: implement `HandlePatchMember` in `server/handlers.go` or new `server/patch.go`; update `UserClient` interface.
4. Wire router: add `case http.MethodPatch` to `server` HTTP switch and update `internal/httpserver` if needed.
5. Add handler unit tests: `server/handlers_patch_test.go` (use `fakeClient` pattern).
6. Add integration tests: `tests/integration/patch_test.go`, gate by `INTEGRATION_TESTS=true`, re-use Samba compose and harness.
7. Update docs and TODOs: `README.md` and `docs/integration-testing.md` with PATCH usage and test instructions.

### Further Considerations
1. Attribute whitelist: decide allowed attributes (displayName, givenName, mail, telephoneNumber, description, etc.) — reject privileged attributes (objectClass, memberOf, userAccountControl, etc.) unless explicitly supported.
2. Password handling: map password changes to existing `setUnicodePwd` logic and require LDAPS; do not log passwords.
3. CI: integration tests are gated; add optional CI job only if maintainers can provide required runners/secrets.

### Gaps from previous repo analysis (updated)
- Missing exported ldaps modify API (no `ModifyUserAttributes` / `PatchAttributes`).
- Missing `PATCH /v1/member` handler and router case.
- `UserClient` interface lacks a modify method; `internal/httpserver` and `server` need updates.
- No unit tests for generic ModifyRequest building beyond unicodePwd; need tests for add/replace/delete ops.
- Integration tests for PATCH are absent — need to add `tests/integration/patch_test.go`.
- Existing repo scan may have missed any references in recently modified files (`config`, `Makefile`, workflow files). Confirm these files do not already add a modify API before implementing.

### Suggested Low-Risk Implementation Approach & Filenames
- High-level approach:
	1. Add a minimal ldaps modify helper that accepts a DN and a list of simple modify operations (op, attribute, values). Reuse `ldapModifier` for testability.
	2. Add server handler that accepts a well-defined JSON patch model (see below), sanitizes it, maps it to ldaps modify ops, special-cases password -> call `setUnicodePwd` instead of a normal modify (or map "password" to unicodePwd and call `setUnicodePwd`).
	3. Add server and internal interfaces and tests. Wire handler into router by adding `case http.MethodPatch` in the server switch.
	4. Add unit tests that reuse `mockModifier` pattern to assert `ModifyRequest` content and handler tests using `fakeClient`.
	5. Add an integration test that creates a user, patches an attribute and/or password, and verifies via GET.

- Files to add / modify (concrete):
	- Add `ldaps/modify.go` — implements ldaps modify helper functions.
	- Add `ldaps/modify_test.go` — tests modify building and behavior (use `mockModifier` pattern).
	- Modify `server/handlers.go` — add `HandlePatchMember` (or create `server/patch.go` with the handler), and extend `UserClient` interface.
	- Modify `internal/httpserver` — extend `UserClient` interface and router switch to support PATCH.
	- Add `server/handlers_patch_test.go` — add tests for PATCH routing and behavior.
	- Add `tests/integration/patch_test.go` — add integration tests for PATCH flows.

### Precise function signatures (descriptions only)
- ldaps-level add:
	- **Function**: `Client.ModifyUserAttributes`
	- **Signature (describe)**: a method on `Client` that accepts `ctx context.Context`, either a `dn string` or `username string` (or both) and a list/slice of simple modify operations (operation type: "add"/"replace"/"delete", attribute name, `values []string`). Returns `error`.
	- **Behavior**: constructs `ldap.ModifyRequest`, for each op apply changes, then calls `conn.Modify(req)` using a `ldapModifier`-typed connection; propagate errors; for password operations prefer `setUnicodePwd` for AD unicodePwd semantics.

- Modify operation descriptor (type definition suggestion)
	- **Type**: `ModifyOperation` (internal/ldaps)
	- **Fields (describe)**: `Op` ("add"|"replace"|"delete"), `Attr` (string), `Values` ([]string)

- server-level handler:
	- **Function**: `HandlePatchMember`
	- **Signature (describe)**: `func HandlePatchMember(client UserClient, w http.ResponseWriter, r *http.Request) error`
	- **Request model**: JSON body like `{"username":"alice","changes":[{"op":"replace","attr":"displayName","values":["Alice Q"]}]}`.
	- **Behavior**:
		- Parse and limit body (similar to POST).
		- Validate `username` present and non-empty.
		- Sanitize username (reuse existing sanitization).
		- For each change: validate op is one of add/replace/delete, attribute name allowed (whitelist), values non-empty as required.
		- If attribute maps to password (e.g., `password`), call ldaps `setUnicodePwd` path (via new `ModifyUserAttributes` or directly call `setUnicodePwd` through exported wrapper).
		- Otherwise, map to ldaps modify ops and call `Client.ModifyUserAttributes`.
		- Return appropriate HTTP codes: 200 on success (or 204), 400 on validation error, 500 on backend error (translated by `makeAppHandler`).

- UserClient interface update:
	- **Description**: extend `UserClient` (and `internal/httpserver.UserClient`) to include:
		- `ModifyUserAttributes(ctx context.Context, username string, ops []ModifyOperation) error`
	- Or: `ModifyMember(ctx context.Context, username string, ops []ModifyOperation) error`

- Mock interface required for unit tests:
	- For ldaps tests: reuse `mockModifier` from `ldaps/unicode_pwd_test.go` but ensure it can capture and inspect `ModifyRequest` for multiple ops.
	- For server handler tests: extend `fakeClient` with `modifyUserAttributes` func field to allow asserting parameters and controlling errors.

### Unit tests to add (names & assertions)
- `ldaps/modify_test.go`:
	- `TestModifyUserAttributes_BuildsModifyRequest` — using `mockModifier`, assert DN, number of Changes, each `Change.Modification.Type` and `Vals` match expected op/attr/values.
	- `TestModifyUserAttributes_PasswordUsesUnicodePwd` — calling Modify with attribute "password" should invoke `setUnicodePwd` path (or result in `unicodePwd` replace with encoded bytes). Use `mockModifier` to inspect replacement bytes and match expected encoding.
	- `TestModifyUserAttributes_PropagatesErrors` — when `mockModifier.Modify` returns error, the method returns error.

- `server/handlers_patch_test.go`:
	- `TestHandlePatchMember_MissingUsername` — send JSON without username -> 400 and error message.
	- `TestHandlePatchMember_InvalidJSON` — malformed JSON -> 400.
	- `TestHandlePatchMember_InvalidChangeOp` — invalid op -> 400.
	- `TestHandlePatchMember_Success` — fakeUserClient.modify called with sanitized username and ops; handler returns 200 (or 204) and correct JSON.
	- `TestHandlePatchMember_BackendError` — fakeUserClient.modify returns error -> handler returns error which translator maps to 500.

- Integration tests:
	- `TestPatchMemberIntegration_UpdateDisplayName` — create user, PATCH displayName, GET user and assert updated displayName.
	- `TestPatchMemberIntegration_ChangePassword` — create user, PATCH password, then verify bind-as-user succeeds (use existing integration harness patterns); if bind-as-user not possible, adapt to verify via an observable side-effect.

### Router wiring
- File: `server/handlers.go`:
	- In the existing HTTP method switch add:
		- `case http.MethodPatch: return HandlePatchMember(client, w, r)`
	- Update `UserClient` interface type in this file to include the new modify signature.

### Validation & Security considerations (short)
- Whitelist attributes allowed to be patched to prevent arbitrary schema changes.
- For password changes require LDAPS and appropriate authentication/authorization. Ensure logs do not leak password content.
- Use existing `setUnicodePwd` helper for AD compatibility.

### Next steps (concrete)
1. Implement `ldaps.ModifyUserAttributes` with `ModifyOperation` descriptor and unit tests using `mockModifier`.
2. Add `HandlePatchMember` and wire into router; add handler unit tests.
3. Add integration test that exercises attribute update and password change using the Samba AD SUT.
4. Run `go test ./...` locally and run integration tests with docker compose.
