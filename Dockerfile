
FROM rust:latest AS clip-builder

ARG DEBIAN_FRONTEND=noninteractive
RUN --mount=type=cache,target=/var/lib/apt/lists \
    --mount=type=cache,target=/var/cache/apt \
    apt-get update && apt-get install -y protobuf-compiler

WORKDIR /src
COPY ./services/clip /src
COPY ./api/proto/clip /src/proto
RUN --mount=type=cache,target=/src/target \
    cargo build --release && mv /src/target/release/clip-service /src/clip-service

#--------------------------------
FROM gcr.io/distroless/cc-debian12 AS clip-service
COPY --from=clip-builder /src/clip-service /
CMD ["./clip-service"]

#--------------------------------
FROM golang:1.23 AS migration-builder

RUN --mount=type=cache,target=/root/go/pkg/mod \
    go install github.com/jackc/tern/v2@latest

#--------------------------------
FROM gcr.io/distroless/cc-debian12 AS migration

WORKDIR /migrations
COPY --from=migration-builder /go/bin/tern /tern
COPY ./migrations /migrations

CMD ["/tern", "migrate"]

#--------------------------------
FROM golang:1.23 AS aio-builder

WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/root/go/pkg/mod \
    go build -o aio-service ./cmd/aiosvc/main.go

#--------------------------------
FROM gcr.io/distroless/cc-debian12 AS aio-service
COPY --from=aio-builder /src/aio-service /
CMD ["./aio-service"]
