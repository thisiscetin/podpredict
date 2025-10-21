# 1) Build stage
FROM golang:1.25 AS builder
WORKDIR /src

# Cache-friendly: copy deps first, then download
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest and build a static binary
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false \
    -ldflags="-s -w" \
    -o /out/server ./cmd/server

# 2) Runtime stage (non-root, CA certs included)
FROM gcr.io/distroless/base-debian12:nonroot

# Optional metadata (can be set via --build-arg)
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

USER nonroot:nonroot
EXPOSE 7000
ENTRYPOINT ["/app/server"]
