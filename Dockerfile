# syntax=docker/dockerfile:1.7

# 1) Build stage
FROM golang:1.25 as builder

WORKDIR /src

# Copy module files first to maximize layer caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Now copy the rest of the source
COPY . .

# Build args for reproducible builds and optional versioning
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE

# Produce a static, trimmed binary
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -buildvcs=false \
      -ldflags="-s -w" \
      -o /out/server ./cmd/server

# 2) Runtime stage
# Use distroless base (has CA certs, runs as non-root)
FROM gcr.io/distroless/base-debian12:nonroot

# Metadata (OCI labels)
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE
LABEL org.opencontainers.image.title="podpredict" \
      org.opencontainers.image.description="Pod prediction API" \
      org.opencontainers.image.url="https://github.com/thisiscetin/podpredict" \
      org.opencontainers.image.source="https://github.com/thisiscetin/podpredict" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}"

WORKDIR /app
COPY --from=builder /out/server /app/server

# Non-root by default in this image
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]