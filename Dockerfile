ARG GO_VERSION=1.15.2

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o ./app ./main.go

FROM alpine:latest

ENV CONFIG=/config
ENV DATA=/assets
ENV UID=998
ENV PID=100
ENV PUID=1000
ENV PGID=1000
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
COPY start.sh .
RUN chmod +x start.sh
RUN apk update && apk add su-exec && rm -rf /var/cache/apk/*s

ENTRYPOINT ["./start.sh"]