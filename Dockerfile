ARG GO_VERSION=1.25.0

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
# mattn/go-sqlite3's cgo shim references pread64/pwrite64/off64_t, which
# musl only exposes when _LARGEFILE64_SOURCE is defined - without this,
# the build fails on current Alpine base images.
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN go build -o ./app ./main.go

FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/akhilrex/podgrab"

ENV CONFIG=/config
ENV DATA=/assets
ENV UID=998
ENV PID=100
ENV GIN_MODE=release
VOLUME ["/config", "/assets"]
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN mkdir -p /config; \
    mkdir -p /assets; \
    mkdir -p /api

RUN chmod 777 /config; \
    chmod 777 /assets

WORKDIR /api
COPY --from=builder /api/app .
COPY client ./client
COPY webassets ./webassets

EXPOSE 8080

ENTRYPOINT ["./app"]