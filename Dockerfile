# Dockerfile - multi-stage build for Goberus
# Stage 1: build binary
FROM golang:1.23 AS builder

# Set working dir
WORKDIR /src

# Copy go.mod/go.sum first to download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the code
COPY . .

# Build static binary
# Disable CGO to produce a static-like binary (helps portability). If you require CGO for TLS or platform-specific behavior, adjust accordingly.
ENV CGO_ENABLED=0
# Allow overriding target OS/ARCH via build-args or let buildx set them automatically when using --platform
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w -X main.Version=$(cat /src/VERSION) -X main.Commit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" -o /goberus ./cmd/goberus

# Stage 2: small runtime image
FROM alpine:3.18 AS runtime

# Install CA certificates (for LDAPS trust)
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -S goberus && adduser -S -G goberus goberus
WORKDIR /app

# Copy binary from builder
COPY --from=builder /goberus /app/goberus
# Make the binary owned by the non-root user
RUN chown goberus:goberus /app/goberus
USER goberus

# Tell the application where to find the CA bundle (optional; app also respects env at runtime)
ENV LDAP_CA_CERT="/etc/ssl/certs/goberus-ca.pem"

# Expose the default port
EXPOSE 8080

# Default command
ENTRYPOINT ["/app/goberus"]
CMD []
