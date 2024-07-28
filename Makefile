VERSION ?= dev
IMAGE ?= ghcr.io/raymanovg/kvstore

.PHONY: build
build:
	go build -o ./bin/http -ldflags '-X main.Version=${VERSION}' ./cmd/http

.PHONY: build-img
build-img:
	docker build -t ${IMAGE}:${VERSION} --build-arg Version=${VERSION} .

.PHONY: run-img
run-img:
	docker run -it --rm --name kvstore -v $(pwd)/config:/opt/kvstore/config ${IMAGE}:${VERSION}

.PHONY: push-img
push-img:
	docker push ${IMAGE}:${VERSION}

.PHONY: build-push
build-push: build-img push-img

.PHONY: run
run:
	./bin/http

.PHONY: brun
brun: build
	./bin/http