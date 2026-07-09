BINARY     := server
MAIN       := .
GOFLAGS    := -trimpath
IMAGE      := local-sqs
IMAGE_TAG  ?= latest

.PHONY: all build test lint fmt clean docker-build docker-push docker-run

all: fmt lint test build

build:
	go build $(GOFLAGS) -o bin/$(BINARY) $(MAIN)

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin/

docker-build:
	docker build -t $(IMAGE):$(IMAGE_TAG) .

docker-push:
	docker push $(IMAGE):$(IMAGE_TAG)

docker-run:
	docker run --rm $(IMAGE):$(IMAGE_TAG)
