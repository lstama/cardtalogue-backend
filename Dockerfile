# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates build-essential pkg-config \
    && if [ "$TARGETARCH" = "arm64" ]; then apt-get install -y --no-install-recommends gcc-aarch64-linux-gnu libc6-dev-arm64-cross g++-aarch64-linux-gnu; fi \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN if [ "$TARGETARCH" = "arm64" ]; then \
      export CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++; \
    else \
      export CC=gcc CXX=g++; \
    fi; \
    CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/cardtalogue-backend ./cmd/api

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/cardtalogue-backend /app/cardtalogue-backend
COPY migrations /app/migrations
COPY uploads /app/uploads

RUN mkdir -p /app/data /app/uploads/media

EXPOSE 8080

ENV PORT=8080 \
    DATABASE_DRIVER=sqlite \
    DATABASE_URL=file:/app/data/cardtalogue.db \
    MIGRATIONS_PATH=file:///app/migrations \
    MEDIA_UPLOAD_PATH=/app/uploads/media

CMD ["/app/cardtalogue-backend"]
