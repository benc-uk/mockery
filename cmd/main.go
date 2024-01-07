package main

// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// Mockery - Main entry point
// ----------------------------------------------------------------------------

import (
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
}

var config = Config{
	specFile: "",
	port:     8000,
	logLevel: slog.LevelInfo,
}

const contentType = "application/json"

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

	// Command line flags
	var levelString string
	flag.StringVar(&config.specFile, "file", "", "OpenAPI spec file in JSON or YAML format. REQUIRED")
	flag.StringVar(&config.specFile, "f", "", "OpenAPI spec file in JSON or YAML format. REQUIRED")
	flag.IntVar(&config.port, "port", 8000, "Port to run mock server on")
	flag.StringVar(&levelString, "log-level", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	levelString = strings.ToLower(levelString)
	if levelString == "debug" {
		config.logLevel = slog.LevelDebug
	} else if levelString == "info" {
		config.logLevel = slog.LevelInfo
	} else if levelString == "warn" {
		config.logLevel = slog.LevelWarn
	} else if levelString == "error" {
		config.logLevel = slog.LevelError
	}

	if config.logLevel != slog.LevelInfo {
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level: config.logLevel,
		}))
	}

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

	r := chi.NewRouter()

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
	r.Use(cors.Handler)

	// Add server headers
	r.Use(middleware.SetHeader("Server", fmt.Sprintf("Mockery: %s v%s", spec.Info.Title, spec.Info.Version)))

	// Loop over all paths
	for path, pathSpec := range spec.Paths {
		if path[:1] != "/" {
			continue
		}

		fullPath := basePath + path
		if pathSpec.isGet() {
			logger.Info("🔵 Adding GET route", slog.Any("path", fullPath))
			r.Get(fullPath, createResponseHandler(pathSpec.Get))
		}

		if pathSpec.isPost() {
			logger.Info("🟢 Adding POST route", slog.Any("path", fullPath))
			r.Post(fullPath, createResponseHandler(pathSpec.Post))
		}

		if pathSpec.isPut() {
			logger.Info("🟠 Adding PUT route", slog.Any("path", fullPath))
			r.Put(fullPath, createResponseHandler(pathSpec.Put))
		}

		if pathSpec.isDelete() {
			logger.Info("🔴 Adding DELETE route", slog.Any("path", fullPath))
			r.Delete(fullPath, createResponseHandler(pathSpec.Delete))
		}
	}

	logger.Warn("Mockery server started", slog.Any("port", config.port))

	// Create custom server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
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
