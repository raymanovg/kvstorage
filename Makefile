VERSION ?= dev

.PHONY: build
build:
	go build -o ./bin/http -ldflags '-X main.Version=${VERSION}' ./cmd/http

run:
	./bin/http

brun: build
	./bin/http