# Mockery

Mockery is a simple tool which runs a HTTP API listener to accept requests based on an Open API Specification (OAS) also known as Swagger. It will parse the provided OAS document and discover paths, responses etc and configure handlers to respond accordingly. Currently it supports v2 of Swagger/OAS.

It can be use to act as mock or placeholder server for testing, mocking, or other uses cases when the real API endpoint is not available.

It goes beyond providing simple empty HTTP responses, and will use any examples discovered in the OAS to provide a payload repsonse back, obviously these responses are static, however they do increase the usefulness of the API tremendously.

![screen shot](./etc/screenshot.png)

# Install & Run

## Download Binary


## Run From Container

A container image is available on GitHub. You will need to mount ot inject the directory where your OAS spec file is located and supply that as an arguement when running, for example:

```bash
docker run -v ./some_directory:/specs \
 -p 8000:8000 \
 ghcr.io/benc-uk/mockery:latest -f /specs/nanomon.json
```

## Go Install 

Install from source if you have Go on your machine

```bash
go install github.com/benc-uk/mockery/cmd@latest
mv $(go env GOPATH)/bin/cmd ~/.local/bin/mockery
```

# Usage

```text
$ mockery
  -f string
        OpenAPI spec file in JSON format. REQUIRED
  -file string
        OpenAPI spec file in JSON format. REQUIRED
  -log-level string
        Log level: debug, info, warn, error (default "info")
  -port int
        Port to run mock server on (default 8000)
```

You must provide an OpenAPI spec file with either `-file` or `-f`. By default it will start and listen on port 8000

## Using Container

Also yes

# Response Handling Logic

## Developer Guide