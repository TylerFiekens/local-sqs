BINARY     := server
MAIN       := .
GOFLAGS    := -trimpath
IMAGE      := local-sqs
IMAGE_TAG  ?= latest

.PHONY: all build test lint fmt clean docker-build docker-push docker-run

all: fmt lint test build

run:
	go run $(GOFLAGS) $(MAIN)

build:
	go build $(GOFLAGS) -o bin/$(BINARY) $(MAIN)

test:
	go test ./... -race

test-coverage:
	go test ./... -race -cover

lint:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin/

check:
	make fmt
	make lint
	make build
	make test-coverage

docker-build:
	docker build -t $(IMAGE):$(IMAGE_TAG) .

docker-push:
	docker push $(IMAGE):$(IMAGE_TAG)

docker-run:
	docker run --rm $(IMAGE):$(IMAGE_TAG)
