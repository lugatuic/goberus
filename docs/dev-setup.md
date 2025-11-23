# Development setup

This document walks through the common commands for running and building Goberus locally.

## Quick start — run locally
1. Prerequisites
   - Go 1.21+
   - Network access to your Active Directory on LDAPS (TCP 636)

2. Set environment variables (example)
```bash
export BIND_ADDR="8080"
export LDAP_ADDR="ad.example.local:636"            # host:port for LDAPS
export LDAP_BASE_DN="DC=example,DC=local"
export LDAP_BIND_DN="CN=svc-goberus,OU=svc,DC=example,DC=local"
export LDAP_BIND_PASSWORD="supersecret"
export LDAP_SKIP_VERIFY="false"                    # set true only for dev testing
# Or better: provide CA cert that signed the AD server cert:
# export LDAP_CA_CERT="/path/to/ca.pem"
```

3. Build and run
```bash
go mod tidy
go build -o goberus ./...
./goberus
```

4. Query the service
```bash
curl 'http://localhost:8080/v1/member?username=jdoe' | jq .
# or with UPN:
curl 'http://localhost:8080/v1/member?username=jdoe@example.local' | jq .

curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"username":"testuser","password":"S3cureP@ss"}' \
  http://localhost:8080/v1/member | jq .
```

## Docker — build and run
The Dockerfile uses a multi-stage build to produce a minimal runtime image.

Build:
```bash
docker build -t lugatuic/goberus:latest .
```

Run (example):
```bash
docker run --rm -p 8080:8080 \
  -e BIND_ADDR=":8080" \
  -e LDAP_ADDR="ad.example.local:636" \
  -e LDAP_BASE_DN="DC=example,DC=local" \
  -e LDAP_BIND_DN="CN=svc-goberus,OU=svc,DC=example,DC=local" \
  -e LDAP_BIND_PASSWORD="supersecret" \
  -e LDAP_SKIP_VERIFY="false" \
  lugatuic/goberus:latest
```

To provide a CA certificate file inside the container, mount it and set `LDAP_CA_CERT`:
```bash
docker run --rm -p 8080:8080 \
  -v /local/path/ca.pem:/etc/ssl/certs/goberus-ca.pem:ro \
  -e LDAP_CA_CERT="/etc/ssl/certs/goberus-ca.pem" \
  ... lugatuic/goberus:latest
```

## CI tip
If you store your CA PEM in a GitHub Actions secret (recommended for private/internal CAs), write it to `ad_chain.pem` during the workflow and build the image. The Dockerfile copies that file to `/etc/ssl/certs/goberus-ca.pem` and sets `LDAP_CA_CERT`.

Example step:
```yaml
- name: Build image (write CA from secret)
  run: |
    echo "$GOBERUS_CA_PEM" > ad_chain.pem
    docker build --build-arg TARGETARCH=amd64 --build-arg TARGETOS=linux -t lugatuic/goberus:latest .
  env:
    GOBERUS_CA_PEM: ${{ secrets.GOBERUS_CA_PEM }}
```

> **Notes**
> - Do NOT commit `ad_chain.pem` to the repository — it is ignored by `.gitignore`.
> - The workflow writes the secret only during the run and uses it to build the image, keeping sensitive data out of the repo.
> - If you prefer not to bake the CA into the image, mount it at runtime instead (see previous section).

## Environment variables reference
- `BIND_ADDR` — HTTP listen address (default `:8080`)
- `LDAP_ADDR` — LDAPS address, host:port (required)
- `LDAP_BASE_DN` — base DN for searches (required)
- `LDAP_BIND_DN` — optional service DN used for searches/modify (recommended)
- `LDAP_BIND_PASSWORD` — password for `LDAP_BIND_DN`
- `LDAP_SKIP_VERIFY` — set to `true` to skip TLS verification (development only)
- `LDAP_CA_CERT` — path to a CA PEM file used to verify the LDAPS server cert

## Behavior & notes
- Authentication: the current implementation prefers bind-as-user for authentication; only the read/search endpoint (`/v1/member`) and the POST `/v1/member` user creation endpoint are exposed.
- Active Directory password operations run over LDAPS using AD's `unicodePwd` behavior when creating users (`ldaps.AddUser` now calls `setUnicodePwd` and `enableAccount`).
- TLS: do not use `LDAP_SKIP_VERIFY=true` in production. Provide a CA via `LDAP_CA_CERT` or trust a CA that already exists in the container.

## Troubleshooting
- `x509: certificate signed by unknown authority`: provide `LDAP_CA_CERT` or ensure the CA is trusted.
- `Bind failed`: ensure `LDAP_BIND_DN` is a full DN and the password is correct; validate with `ldapsearch` if necessary.
- No entries found: verify `LDAP_BASE_DN` and try searching by different username formats (sAMAccountName, UPN).

## Testing
- `go test ./tests/server -run TestHandleGetMember`
- `go test ./tests/server -run TestHandleCreateMember`
- `go test ./tests/server -run TestSanitizeUserIntegration`
- `go test ./...`
