# Goberus

Goberus is a minimal LDAPS middleware implementing a get_member_info endpoint for Active Directory / LDAP.  
This repository contains a small Go service that connects to an LDAPS server, searches for a user (by userPrincipalName or sAMAccountName), and returns a JSON representation of select attributes.

This repo is intentionally minimal so you can validate functionality before we extend it further (password operations, API auth, pooling, tests, etc.).

Status
- Initial implementation: GET /v1/member?username=<username>
- LDAPS client dials and binds per request (simple, robust for testing)
- TLS verification is configurable (recommended to provide a CA cert)

Quick start — run locally
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
```

Docker — build and run
A multi-stage Dockerfile is provided to build a small runtime image.

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

If you need to provide a CA certificate file to verify the LDAPS server certificate, mount it into the container and set `LDAP_CA_CERT` to the path inside the container:
```bash
docker run --rm -p 8080:8080 \
  -v /local/path/ca.pem:/etc/ssl/certs/goberus-ca.pem:ro \
  -e LDAP_CA_CERT="/etc/ssl/certs/goberus-ca.pem" \
  ... lugatuic/goberus:latest
```

CI tip — bake CA into the image from a secret (GitHub Actions example)
If you store your CA PEM as a repository secret (recommended for private/internal CAs), write it to `ad_chain.pem` in the workflow and build the image. The `Dockerfile` will copy `ad_chain.pem` into `/etc/ssl/certs/goberus-ca.pem` and set `LDAP_CA_CERT`.

Example GitHub Actions step (assumes secret name GOBERUS_CA_PEM):

```yaml
- name: Build image (write CA from secret)
  run: |
    echo "$GOBERUS_CA_PEM" > ad_chain.pem
    docker build --build-arg TARGETARCH=amd64 --build-arg TARGETOS=linux -t lugatuic/goberus:latest .
  env:
    GOBERUS_CA_PEM: ${{ secrets.GOBERUS_CA_PEM }}
```

Notes:
- Do NOT commit `ad_chain.pem` to the repository — it's added to `.gitignore` by default.
- The workflow writes the secret to a file only during the run and uses it to create the image. This keeps the private CA material out of the repo while allowing the image to include the CA at build time.
- If you prefer not to bake the CA into the image, mount it at runtime instead (see example above).

Environment variables reference
- BIND_ADDR — HTTP listen address (default `:8080`)
- LDAP_ADDR — LDAPS address, host:port (required)
- LDAP_BASE_DN — base DN for searches (required)
- LDAP_BIND_DN — optional service DN used for searches/modify (recommended)
- LDAP_BIND_PASSWORD — password for LDAP_BIND_DN
- LDAP_SKIP_VERIFY — "true" to skip TLS verification (development only)
- LDAP_CA_CERT — path to CA PEM file used to verify the LDAPS server cert

Behavior & notes
- Authentication: the service currently implements bind-as-user for authentication in the original design; the only exposed endpoint initially is a read/search endpoint (`/v1/member`). We will add API auth (JWT/API keys) later.
- Active Directory specifics: setting/changing passwords in AD requires LDAPS and AD's unicodePwd behavior (UTF-16LE quoted string). That is not implemented yet.
- TLS: for production, do not use `LDAP_SKIP_VERIFY=true`. Provide a CA via `LDAP_CA_CERT` or ensure the directory's cert chains to a trusted root in the container.

Troubleshooting
- "x509: certificate signed by unknown authority": provide LDAP_CA_CERT or ensure the CA is installed/trusted.
- "Bind failed": ensure LDAP_BIND_DN is a full DN and the password is correct; try a manual ldapsearch to validate credentials.
- No entries found: check LDAP_BASE_DN is correct and try searching by different username formats (sAMAccountName and UPN).

Next steps (after you verify this works)
- Add API authentication and rate limiting
- Implement connection pooling/reconnect semantics
- Implement AD-safe password set/change (unicodePwd handling over LDAPS)
- Add unit tests + integration tests (Docker-compose with Samba AD)

License
- Add your preferred license; none is included by default.
