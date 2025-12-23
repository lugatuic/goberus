# DELETE /v1/member — User deletion support

## Overview
Add `DELETE /v1/member` endpoint for deleting users from Active Directory.

## Implementation tasks

### 1. ldaps-level delete support (`ldaps/delete.go` + tests)
- [ ] Implement `Client.DeleteUser(ctx context.Context, username string) error`
  - Resolves username to DN via GetMemberInfo
  - Calls ldap.Del(dn)
  - Returns error if user not found or deletion fails

### 2. HTTP DELETE handler (`server/handlers.go` or `server/delete.go`)
- [ ] Implement `HandleDeleteMember(client UserClient, w http.ResponseWriter, r *http.Request) error`
  - Extract username from query param or request body
  - Validate username
  - Call `client.DeleteUser()`
  - Return 204 No Content on success; 400/404/500 on error
- [ ] Unit tests: missing username → 400, user not found → 404, success → 204, backend error → 500

### 3. Router wiring
- [ ] Update `server/handlers.go` UserClient interface: add `DeleteUser()` method
- [ ] Update `internal/httpserver` UserClient interface: add `DeleteUser()` method
- [ ] Wire `HandleDeleteMember` into HTTP router: add `case http.MethodDelete` in handler switch

### 4. Integration tests (`tests/integration/delete_test.go`)
- [ ] TestDeleteMemberIntegration_Success: create user → DELETE → GET returns 404
- [ ] TestDeleteMemberIntegration_NotFound: DELETE non-existent user → 404
- [ ] Gate with `INTEGRATION_TESTS=true`

### 5. Documentation
- [ ] Update `README.md`: document DELETE endpoint
- [ ] Update `docs/integration-testing.md`: note DELETE behavior

## Testing checklist
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (all unit tests)
- [ ] `INTEGRATION_TESTS=true go test ./tests/integration -v` passes with Samba AD running

## Files to create/modify
| File | Action | Notes |
|------|--------|-------|
| `ldaps/delete.go` | Create | DeleteUser method |
| `ldaps/delete_test.go` | Create | Unit tests using mockModifier |
| `server/handlers.go` or `server/delete.go` | Create/Modify | HandleDeleteMember + update UserClient interface |
| `internal/httpserver` | Modify | Update UserClient interface + router case |
| `tests/integration/delete_test.go` | Create | Integration tests |
| `README.md` | Modify | Document DELETE endpoint |
| `docs/integration-testing.md` | Modify | Add DELETE behavior notes |

## Security notes
- Enforce LDAPS/TLS for all DELETE operations
- Consider audit logging for deletions
- May require elevated permissions in AD
