# 🎭 Mockery

Mockery is a tool for creating a mock API from a Open API Specification (OAS or Swagger), it runs a HTTP listener accepting requests based on the provided spec. It will parse the OAS document and discover paths, responses, schemas etc and configure handlers to respond accordingly. Currently it supports v2 of Swagger/OAS. It is written in Go and uses the [Chi router & mux](https://github.com/go-chi/chi)

It can be use to act as mock or placeholder server for testing, mocking, or other uses cases when the real API endpoint is not available.

It goes beyond providing simple empty HTTP responses, and will use any examples discovered in the OAS to provide a JSON payload response back, obviously these responses are static, however they do increase the usefulness of the API tremendously.

[![CI Pipeline](https://github.com/benc-uk/mockery/actions/workflows/ci-build.yml/badge.svg)](https://github.com/benc-uk/mockery/actions/workflows/ci-build.yml)

![](https://img.shields.io/github/last-commit/benc-uk/mockery) ![](https://img.shields.io/github/release-date/benc-uk/mockery) ![](https://img.shields.io/github/v/release/benc-uk/mockery) ![](https://img.shields.io/github/commit-activity/m/benc-uk/mockery)

![screen shot](./etc/screenshot.png)

# 💾 Install & Run

## Download Binary

Binaries are available as GitHub releases https://github.com/benc-uk/mockery/releases/latest

Quick download on Linux

```bash
wget https://github.com/benc-uk/mockery/releases/download/0.0.2/mockery-linux -O mockery
chmod +x ./mockery
```

## Run From Container

A container image is available on GitHub. You will need to mount ot inject the directory where your OAS spec file is located and supply that as an argument when running, for example:

```bash
docker run -v ./some_directory:/specs \
 -p 8000:8000 \
 ghcr.io/benc-uk/mockery:latest -f /specs/swagger.json
```

## Go Install

Install from source if you have Go on your machine

```bash
go install github.com/benc-uk/mockery/cmd@latest
mv $(go env GOPATH)/bin/cmd ~/.local/bin/mockery
```

# 💻 Usage

Mockery is a command line tool, with only a handful of arguments. You must provide an OpenAPI spec file with either `-file` or `-f`. By default the services will start and listen on port 8000

```
  -api-key string
        Enable API key authentication
  -cert-path string
        Path to directory wth cert.pem & key.pem to enable TLS
  -f string
        OpenAPI spec file in JSON or YAML format. REQUIRED
  -file string
        OpenAPI spec file in JSON or YAML format. REQUIRED
  -log-level string
        Log level: debug, info, warn, error (default "info")
  -port int
        Port to run mock server on (default 8000)
```

## Config

Configuration can be provided as command line arguments as described above, in addition environmental variables can also be set & used, these will override any set on the command line 

| Variable name | Matching argument |
| ------------- | ----------------- |
| PORT          | `-port`           |
| SPEC_FILE     | `-file`           |
| LOG_LEVEL     | `-log-level`      |
| API_KEY       | `-api-key`        |
| CERT_PATH     | `-cert-path`      |

# 🧩 Response Handling Logic

The OAS spec is parsed and used with the following logic:

- Routes are taken from the `paths` section, with matching operations, e.g. `GET` & `POST` etc. a HTTP handler is created for each path and method.
- Path parameters enclosed in `{}` like `/api/orders/{orderId}` are matched as part of the route.
- The `responses` section is scanned for a response status code, 200 is the default
  - If 200 is not a present in responses, then the first response in the list is used.
  - To get a different response/status supply the `x-mock-response-code` header on the request.
- To create a payload for the response, the selected response object is used as follows:
  - If the response has an `examples` field the `application/json` key is used & returned.
  - Otherwise if the response has a `schema` and this schema has an `example` it is used & returned.
  - Otherwise if the response has a `schema` it is parsed and traversed, the fields `properties`, `items` are used and `$ref` can reference models from the `definitions` section of the spec.
    - If no `example` are found at the field level, a fallback default value for the type is used, e.g. `"string"` or `0` or `false`

# 🧑‍💻 Developer Guide

Pre-reqs

- Go v1.21+
- Linux/bash/make

A makefile acts as the frontend & guide to working locally with the project, running `make install-tools` will install required dev tools locally. The other targets are fairly self explanatory. All Go source is in `cmd/` and single 'main' package for simplicity

```text
$ make

help                 💬 This help message :)
install-tools        🔮 Install dev tools into project .tools directory
lint                 🔍 Lint & format check only, sets exit code on error for CI
lint-fix             📝 Lint & format, attempts to fix errors & modify code
image                📦 Build container image from Dockerfile
push                 📤 Push container image to registry
build                🔨 Build binaries for all platforms
run                  🏃 Test and hotreload the app
clean                🧹 Clean up, remove dev data and files
release              🚀 Release a new version on GitHub
```

## Common Dev Tasks

### Linting

Make sure you have run `make install-tools` first

```bash
make lint
```

### Building & Pushing Images

Set the tag you want to build & push and run

```bash
IMAGE_TAG=x.y.z make image push
```

### Releasing to GitHub

Set the version you want to release and run

```bash
VERSION=x.y.z make release
```
