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

func TestFileParser(t *testing.T) {
	// Create a temporary file with some valid OpenAPI v2 spec
	tempFileJSON, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFileJSON.Name())
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

	tempInvalidFile, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempInvalidFile.Name())
	_, err = tempInvalidFile.Write([]byte(`blah wibble`))
	if err != nil {
		t.Fatal(err)
	}
	tempInvalidFile.Close()

	t.Run("valid_json", func(t *testing.T) {
		// Test parsing a valid file
		_, err = ParseV2Spec(tempFileJSON.Name())
		if err != nil {
			t.Errorf("failed with valid file, got: %v", err)
		}
	})

	// Test parsing a valid YAML file
	t.Run("valid_yaml", func(t *testing.T) {
		_, err = ParseV2Spec(tempFileYAML.Name())
		if err != nil {
			t.Errorf("failed with valid YAML file, got: %v", err)
		}
	})

	// Test parsing a non-existent file
	t.Run("no_file", func(t *testing.T) {
		_, err = ParseV2Spec("non_existent_file.yaml")
		if err == nil {
			t.Error("did not fail with non-existent file")
		}
	})

	// Test parsing an invalid file
	t.Run("invalid_file", func(t *testing.T) {
		_, err = ParseV2Spec(tempInvalidFile.Name())
		if err == nil {
			t.Error("did not fail with invalid file")
		}
	})

}

func TestResponseParser(t *testing.T) {
	emptyResp := Response{}

	respExampleJSON := Response{
		Description: "Response with example",
		Examples: map[string]any{
			"application/json": map[string]any{
				"key1": "anything",
				"key2": 42,
			},
		},
	}

	respExamplePlain := Response{
		Description: "Response with example",
		Examples: map[string]any{
			"text/plain": "Hello, World!",
		},
	}

	// Test parsing an empty response
	t.Run("empty", func(t *testing.T) {
		if emptyResp.parse() != nil {
			t.Error("expected nil data from response.parse()")
		}
	})

	// Test parsing a response with an example
	t.Run("resp_example_json", func(t *testing.T) {
		data := respExampleJSON.parse()
		if data == nil {
			t.Error("expected data from response.parse()")
		}

		// convert to map
		dataMap, ok := data.(map[string]any)
		if !ok {
			t.Error("expected data to be a map")
		}

		// check that the map has the expected keys
		if value, ok := dataMap["key1"]; true {
			if !ok {
				t.Error("expected key 'key1' in data map")
			}

			if value.(string) != "anything" {
				t.Errorf("expected value of 'key1' to be 'anything', got: %v", value)
			}
		}
	})

	t.Run("resp_example_plain", func(t *testing.T) {
		if respExamplePlain.parse() != nil {
			t.Error("expected nil data from response.parse()")
		}
	})
}
