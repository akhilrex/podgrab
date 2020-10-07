ARG GO_VERSION=1.15.2
ENV UID=998
ENV GID=100
ENV CONFIG=/config
ENV DATA=/assets

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

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api
COPY --from=builder /api/app .
COPY client ./client
RUN mkdir /config; \
    mkdir /assets

RUN  groupadd -g ${GID} poduser &&\
    useradd -l -u ${UID} -g poduser poduser &&\
    chown --changes --silent --no-dereference --recursive \
           ${UID}:${GID} \
        /assets \
        /config 

USER poduser

EXPOSE 8080
VOLUME ["/config", "/assets"]
ENTRYPOINT ["./app"]