# local-sqs

A lightweight, in-memory SQS-compatible server for local development and testing.

## Status

All core operations implemented.

## How it works

Implements the [AWS SQS Query API](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/query-interface.html) over HTTP. Clients send `POST /` with `Content-Type: application/x-www-form-urlencoded` and an `Action` field. Responses are SQS-compatible XML.

Queues and messages are held in memory — state is lost on restart.

## Usage

### Go

```sh
go run . -port 9324
```

### Docker

```sh
docker run -p 9324:9324 tylerfiekens/local-sqs:latest
```

### Docker Compose

```yaml
services:
  local-sqs:
    image: tylerfiekens/local-sqs:latest
    ports:
      - "9324:9324"
```

Point your AWS SDK at `http://localhost:9324` with any credentials (not validated).

The OpenAPI specification is served at `http://localhost:9324/openapi.yaml`.

## Implemented operations

| Operation | Status |
|---|---|
| `CreateQueue` | Implemented |
| `DeleteQueue` | Implemented |
| `ListQueues` | Implemented |
| `SendMessage` | Implemented |
| `ReceiveMessage` | Implemented |
| `DeleteMessage` | Implemented |

## Development

Requires Go 1.26+.

```sh
make build          # compile
make test           # run tests with race detector
make check          # fmt + lint + build + test
make docker-build   # build Docker image
```

Pre-built images are published to [Docker Hub](https://hub.docker.com/r/tylerfiekens/local-sqs) on every push to `main`.

## Contributing

1. Fork the repo and create a branch from `main`.
2. Make your changes with tests covering new behaviour.
3. Run `make check` — all checks must pass before opening a PR.
4. Open a pull request against `main` with a clear description of what and why.

## Issues

Found a bug or missing operation? [Open an issue](https://github.com/tylerfiekens/local-sqs/issues/new) with:

- What you did
- What you expected
- What happened (include logs or error output where relevant)
