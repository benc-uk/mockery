# ===================================================================================
# === Stage 1: Build the binary =====================================================
# ===================================================================================
FROM golang:1.21-alpine as go-build

WORKDIR /build

# We need root certs to make HTTPS calls
RUN apk add --no-cache ca-certificates

# Fetch and cache Go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy in Go source files
COPY src/ ./src/

# Now run the build
RUN go build -o mockery ./src

# ================================================================================================
# === Stage 2: Get binary into a lightweight container ===========================================
# ================================================================================================
FROM alpine:3.14
WORKDIR /app 

# Copy the Go server binary
COPY --from=go-build /build/mockery . 

# And root certs too
COPY --from=go-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8000

ENTRYPOINT [ "./mockery" ]