# Multi-stage build for a tiny, secure image
# 1) Build static binary
# 2) Run in minimal Alpine with non-root user

# ---------- Builder ----------
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Speed up `go build` by caching modules
RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a statically-linked binary (no CGO) with optimizations
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64}
RUN go build -trimpath -ldflags "-s -w -extldflags -static" -o /out/mentat .

# ---------- Runtime ----------
FROM alpine:3.20

# Optional: `libcap-utils` allows setting file capabilities if you prefer not to run the container with --cap-add=NET_RAW.
# Note: File capabilities require the image filesystem to support xattrs and may not survive certain copy operations.
# We leave it installed as it's small. You can safely remove if you use Docker/K8s capabilities instead.
RUN apk add --no-cache ca-certificates libcap-utils busybox && update-ca-certificates

# App directory and user
WORKDIR /app

# Copy binary
COPY --from=builder /out/mentat /app/mentat

# Optionally grant the binary permission to create raw sockets for ICMP without container NET_RAW capability
# Uncomment the next line if you prefer file capabilities instead of container capabilities
# RUN setcap cap_net_raw+ep /app/mentat

# Create a non-root user (65532 is the "nonroot" uid used by distroless)
RUN addgroup -S app && adduser -S -G app -u 65532 app
USER 65532:65532

# Export metrics port
EXPOSE 2112

# Basic healthcheck against the metrics endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q -O- http://127.0.0.1:2112/metrics >/dev/null || exit 1

# OCI labels
LABEL org.opencontainers.image.title="mentat" \
      org.opencontainers.image.description="Inter-node latency exporter for Kubernetes (Prometheus)" \
      org.opencontainers.image.source="https://github.com/nathanmartins/mentat" \
      org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["/app/mentat"]