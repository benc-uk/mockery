package main

// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// Mockery - Main entry point
// ----------------------------------------------------------------------------

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/lmittmann/tint"
	"moul.io/banner"
)

type Config struct {
	specFile string
	port     int
	logLevel slog.Level
	apiKey   string
	certPath string
}

const contentType = "application/json"

// Globals, so sue me
var logger *slog.Logger
var spec OpenAPIv2

func init() {
	// Fall back logger, if no config is loaded
	logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level: slog.LevelInfo,
	}))
}

// Main entry point
func main() {
	fmt.Println(banner.Inline("mockery"))

	var config = Config{
		specFile: "",
		port:     8000,
		logLevel: slog.LevelInfo,
		apiKey:   "",
		certPath: "",
	}

	// Populate config from command line flags and environment variables
	config.process()

	if config.specFile == "" {
		logger.Error("No OpenAPI spec file specified, please use -file or -f")
		os.Exit(1)
	}

	// Load spec file
	logger.Info("Will try to load spec document: " + config.specFile)
	var err error
	spec, err = ParseV2Spec(config.specFile)
	if err != nil {
		logger.Error("Failed to parse OpenAPI spec file:", tint.Err(err))
		os.Exit(1)
	}

	// The main router used by the server for all requests
	router := chi.NewRouter()

	// Handle base path
	basePath := spec.BasePath

	// If base path doesn't start with a slash it's malformed
	if basePath == "" || basePath[:1] != "/" {
		logger.Warn("Base path maybe invalid or empty", slog.Any("basePath", basePath))
		basePath = "/"
	}

	if basePath[len(basePath)-1:] == "/" {
		basePath = basePath[:len(basePath)-1]
	}

	// Get some details from the spec
	title := "Untitled API"
	version := "0.0.0"
	if spec.Info.Title != "" {
		title = spec.Info.Title
	}
	if spec.Info.Version != "" {
		version = spec.Info.Version
	}

	logger.Warn("Starting Mockery", slog.Any("title", title), slog.Any("version", version))

	// Ignore *all* CORS, this is a mock server after all
	cors := cors.AllowAll()
	router.Use(cors.Handler)

	// Add server headers
	router.Use(middleware.SetHeader("Server", fmt.Sprintf("Mockery: %s v%s", spec.Info.Title, spec.Info.Version)))

	// Custom not found handler
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		logger.Error("Not found", slog.Any("path", r.URL.Path))
		w.WriteHeader(404)
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mockery - " + title + " v" + version + "\n"))
	})

	// Check for x-api-key header if auth is enabled
	if config.apiKey != "" {
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("x-api-key") == "" {
					logger.Error("Not authorised, missing API key")
					w.WriteHeader(401)
					return
				}

				if r.Header.Get("x-api-key") != config.apiKey {
					logger.Error("Invalid API key")
					w.WriteHeader(401)
					return
				}

				next.ServeHTTP(w, r)
			})
		})
	}

	// Loop over all paths
	for path, pathSpec := range spec.Paths {
		if path[:1] != "/" {
			continue
		}

		fullPath := basePath + path
		if pathSpec.isGet() {
			logger.Info("🔵 Adding GET route", slog.Any("path", fullPath))
			router.Get(fullPath, createResponseHandler(pathSpec.Get))
		}

		if pathSpec.isPost() {
			logger.Info("🟢 Adding POST route", slog.Any("path", fullPath))
			router.Post(fullPath, createResponseHandler(pathSpec.Post))
		}

		if pathSpec.isPut() {
			logger.Info("🟠 Adding PUT route", slog.Any("path", fullPath))
			router.Put(fullPath, createResponseHandler(pathSpec.Put))
		}

		if pathSpec.isDelete() {
			logger.Info("🔴 Adding DELETE route", slog.Any("path", fullPath))
			router.Delete(fullPath, createResponseHandler(pathSpec.Delete))
		}
	}

	useTLS := false

	// Check for TLS cert & key files if certPath is set
	if config.certPath != "" {
		logger.Debug("Enabling TLS, checking cert & key files", slog.Any("certPath", config.certPath))

		useTLS = true

		// Check cert & key files exist
		if _, err := os.Stat(config.certPath + "/cert.pem"); os.IsNotExist(err) {
			logger.Error("cert.pem not found, TLS will be disabled", slog.Any("certPath", config.certPath))

			useTLS = false
		}

		if _, err := os.Stat(config.certPath + "/key.pem"); os.IsNotExist(err) {
			logger.Error("key.pem not found, TLS will be disabled", slog.Any("certPath", config.certPath))

			useTLS = false
		}
	}

	// Create custom server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Warn("Mockery server started", slog.Any("port", config.port), slog.Any("tls", useTLS))

	// If TLS is enabled, start using ListenAndServeTLS
	if useTLS {
		srv.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		if err := srv.ListenAndServeTLS(config.certPath+"/cert.pem", config.certPath+"/key.pem"); err != nil {
			logger.Error("Failed to start TLS server", slog.Any("error", err))
			os.Exit(1)
		}
	}

	// Start regular HTTP listener
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}

// This is the heart of the mocking server, it creates a handler function for a given operation
// The handler function will return a response based on the operation's responses
// And will try to construct a response payload from examples in the spec
func createResponseHandler(op Operation) http.HandlerFunc {
	logger.Debug("   Creating handler", slog.Any("id", op.OperationID), slog.Any("title", op.Description))

	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Request", slog.Any("method", r.Method), slog.Any("path", r.URL.Path),
			slog.Any("id", op.OperationID))

		// Get x-mock-response-code header which allows caller to request a specific response
		requestedCode := r.Header.Get("x-mock-response-code")
		if requestedCode != "" {
			logger.Info("Requested response code", slog.Any("code", requestedCode))
		}

		// Path to discover which response to use
		expectedStatus := 200
		if requestedCode != "" {
			expectedStatus, _ = strconv.Atoi(requestedCode)
		}

		respIndex := strconv.Itoa(expectedStatus)
		statusCode := expectedStatus

		resp, respExists := op.Responses[respIndex]
		if !respExists {
			// No matching response, fall back to first one in map
			logger.Warn("No response matching status, falling back to first response",
				slog.Any("requested_code", requestedCode))

			for respName := range op.Responses {
				respIndex = respName
				statusCode, _ = strconv.Atoi(respName)
				resp = op.Responses[respName]
				break
			}
		}

		// Mutate the response object to add the status code, as a convenience
		resp.StatusCode = statusCode

		// This starts the payload & example discovery process
		payload := resp.parse()

		// Finally return the response with or without payload
		if payload != nil {
			logger.Debug("Returning example payload")
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(statusCode)
			_ = json.NewEncoder(w).Encode(payload)
		} else {
			logger.Warn("No example found, response will be empty", slog.Any("status", respIndex))
			w.WriteHeader(statusCode)
		}
	}
}

// Process command line flags and environment variables to build config
func (c *Config) process() {
	// Command line flags
	var levelString string
	flag.StringVar(&c.specFile, "file", "", "OpenAPI spec file in JSON or YAML format. REQUIRED")
	flag.StringVar(&c.specFile, "f", "", "OpenAPI spec file in JSON or YAML format. REQUIRED")
	flag.IntVar(&c.port, "port", 8000, "Port to run mock server on")
	flag.StringVar(&levelString, "log-level", "info", "Log level: debug, info, warn, error")
	flag.StringVar(&c.apiKey, "api-key", "", "Enable API key authentication")
	flag.StringVar(&c.certPath, "cert-path", "", "Path to directory wth cert.pem & key.pem to enable TLS")
	flag.Parse()

	// Environment variables can override command line flags
	if os.Getenv("SPEC_FILE") != "" {
		c.specFile = os.Getenv("SPEC_FILE")
	}

	if os.Getenv("API_KEY") != "" {
		c.apiKey = os.Getenv("API_KEY")
	}

	if os.Getenv("CERT_PATH") != "" {
		c.certPath = os.Getenv("CERT_PATH")
	}

	if os.Getenv("LOG_LEVEL") != "" {
		levelString = os.Getenv("LOG_LEVEL")
	}

	portEnv := os.Getenv("PORT")
	if portEnv != "" {
		if port, err := strconv.Atoi(portEnv); err == nil {
			c.port = port
		}
	}

	// Print help if no args
	if c.specFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Set log level
	levelString = strings.ToLower(levelString)
	switch levelString {
	case "debug":
		c.logLevel = slog.LevelDebug
	case "info":
		c.logLevel = slog.LevelInfo
	case "warn":
		c.logLevel = slog.LevelWarn
	case "error":
		c.logLevel = slog.LevelError
	default:
		c.logLevel = slog.LevelInfo
	}

	if c.logLevel != slog.LevelInfo {
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level: c.logLevel,
		}))
	}
}
