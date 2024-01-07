# Mockery

Mockery is a simple tool which runs a HTTP API listener to accept requests based on an Open API Specification (OAS) also known as Swagger. It will parse the provided OAS document and discover paths, responses etc and configure handlers to respond accordingly. Currently it supports v2 of Swagger/OAS.

It can be use to act as mock or placeholder server for testing, mocking, or other uses cases when the real API endpoint is not available.

It goes beyond providing simple empty HTTP responses, and will use any examples discovered in the OAS to provide a payload repsonse back, obviously these responses are static, however they do increase the usefulness of the API tremendously.

![screen shot](https://private-user-images.githubusercontent.com/14982936/294769200-31b0dca6-f464-4acb-8b4b-2b6782f46ccb.png?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3MDQ2NDMwNTEsIm5iZiI6MTcwNDY0Mjc1MSwicGF0aCI6Ii8xNDk4MjkzNi8yOTQ3NjkyMDAtMzFiMGRjYTYtZjQ2NC00YWNiLThiNGItMmI2NzgyZjQ2Y2NiLnBuZz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNDAxMDclMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjQwMTA3VDE1NTIzMVomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPTRmMThiMWRiYjU3MDRiNWMxNDE1N2NkNDA4YjA5OWYxYjhiYmFkOTdlNDZmYjI1N2M1OTk1YjAyMmMxYzVjMTImWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0JmFjdG9yX2lkPTAma2V5X2lkPTAmcmVwb19pZD0wIn0.v25mm-Pxw-wNuH6xzg6pYvL5Nr0qHunzadPtEXew1Gw)

# Install

Yes

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