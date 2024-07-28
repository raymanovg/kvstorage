FROM golang:1.22 AS build

ARG Version="dev"

WORKDIR /usr/src/kvstore

COPY go.mod go.sum ./

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/kvstore -ldflags '-X main.Version=${Version}' ./cmd/http

FROM alpine:3.20.1

COPY --from=build /usr/local/bin/kvstore /opt/kvstore/bin/http
COPY config/dev.yml /opt/kvstore/config/dev.yml

WORKDIR /opt/kvstore

CMD ["./bin/http"]