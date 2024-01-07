package main

import (
	"log/slog"
	"os"
	"testing"
)

func init() {
	// Disable most logging during tests
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func TestParseV2Spec(t *testing.T) {
	// Create a temporary file with some valid OpenAPI v2 spec
	tempFileJSON, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tempFileJSON.Name())

	// write a simple JSON spec to the file
	_, err = tempFileJSON.Write([]byte(`
{
	"swagger": "2.0",
	"info": {
		"title": "Test API",
		"version": "1.0.0"
	},
	"paths": {}
}
`))
	if err != nil {
		t.Fatal(err)
	}

	tempFileJSON.Close()

	// Create a temporary file with YAML spec
	tempFileYAML, err := os.CreateTemp("", "*.yaml")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tempFileYAML.Name())

	// write a simple YAML spec to the file
	_, err = tempFileYAML.Write([]byte(`
swagger: "2.0"
info:
	title: "Test API"
	version: "1.0.0"
paths: {}
`))
	if err != nil {
		t.Fatal(err)
	}

	tempFileYAML.Close()

	t.Run("valid_json", func(t *testing.T) {
		// Test parsing a valid file
		_, err = ParseV2Spec(tempFileJSON.Name())
		if err != nil {
			t.Errorf("ParseV2Spec failed with valid file, got: %v", err)
		}
	})

	// Test parsing a valid YAML file
	t.Run("valid_yaml", func(t *testing.T) {
		_, err = ParseV2Spec(tempFileYAML.Name())
		if err != nil {
			t.Errorf("ParseV2Spec failed with valid YAML file, got: %v", err)
		}
	})

	// Test parsing a non-existent file
	t.Run("no_file", func(t *testing.T) {
		_, err = ParseV2Spec("non_existent_file.yaml")
		if err == nil {
			t.Error("ParseV2Spec did not fail with non-existent file")
		}
	})
}

func TestParseResponse(t *testing.T) {
	emptyResp := Response{
		Description: "An empty response",
		Schema:      Schema{},
	}

	respExample := Response{
		Description: "Response with example",
		Examples: map[string]any{
			"application/json": map[string]any{
				"foo":    "bar",
				"number": 42,
			},
		},
	}

	schemaExample := Response{
		Description: "Schema with example",
		Schema: Schema{
			Example: "Hello, World!",
		},
	}

	// Test parsing an empty response
	t.Run("empty", func(t *testing.T) {
		data := emptyResp.parse()
		if data != nil {
			t.Error("parseResponse failed with empty response")
		}
	})

	// Test parsing a response with an example
	t.Run("empty", func(t *testing.T) {
		data := respExample.parse()
		if data == nil {
			t.Error("parseResponse failed with example")
		}
	})

	// Test parsing a schema level example
	t.Run("empty", func(t *testing.T) {
		data := schemaExample.parse()
		if data == nil {
			t.Error("parseResponse failed with schema + example")
		}
	})
}
