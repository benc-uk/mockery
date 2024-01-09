package main

// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// Mockery - Core parser functions
// ----------------------------------------------------------------------------

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

// ParseV2Spec parses an OpenAPI v2 spec file
func ParseV2Spec(filePath string) (OpenAPIv2, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return OpenAPIv2{}, err
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return OpenAPIv2{}, err
	}

	var openAPIv2 OpenAPIv2

	// Handle YAML format as well
	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		err = yaml.Unmarshal(data, &openAPIv2)
		return openAPIv2, err
	}

	err = json.Unmarshal(data, &openAPIv2)
	return openAPIv2, err
}

// Parsing a schema is a bit of a nightmare, this is the entry point
func (s Schema) parse() interface{} {
	logger.Debug("Parsing schema", slog.Any("schema", s))

	if s.isEmpty() {
		return nil
	}

	// Simple case: Schema has example object
	if s.Example != nil {
		return s.Example
	}

	var ref string
	if s.isRef() {
		ref = s.Ref
	}

	// Special case for array of references
	if s.Items.Ref != "" && s.Type == "array" {
		ref = s.Items.Ref
	}

	// Another special case for array or object with items + properties
	if s.Items.Properties != nil && (s.Type == "array" || s.Items.Type == "object") {
		if s.Type == "array" {
			return []interface{}{parseProperties(s.Items.Properties)}
		}
		return parseProperties(s.Items.Properties)
	}

	// Resolve references, this is a bit of a hack but seems ok
	if ref != "" {
		// Split ref string into parts
		refParts := strings.Split(ref, "/")
		// last part is the model name
		modelName := refParts[len(refParts)-1]

		// Get model definition
		referencedSchema, defExists := spec.Definitions[modelName]
		if !defExists {
			return nil
		}

		// Parse definition
		logger.Info("Parsing model", slog.Any("name", modelName))
		parsedSchema := referencedSchema.parse()

		// If it's an array, return an array of the parsed schema
		if s.Type == "array" {
			return []interface{}{parsedSchema}
		}

		return parsedSchema
	}

	// Special case for additionalProperties weirdness
	if s.AdditionalProperties != nil {
		_, isBool := s.AdditionalProperties.(bool)
		if isBool {
			return map[string]interface{}{
				"key": "value",
			}
		}

		if s.AdditionalProperties.(map[string]interface{})["type"] == "string" {
			return map[string]interface{}{
				"key 1": "value 1",
				"key 2": "value 2",
			}
		}
		if s.AdditionalProperties.(map[string]interface{})["type"] == "integer" {
			return map[string]interface{}{
				"key 1": 0,
				"key 2": 1,
			}
		}
	}

	// Schema might just be a bag of properties, the OAS spec is a nightmare
	if s.Properties != nil {
		return parseProperties(s.Properties)
	}

	// If we get here, we don't know what to do
	return nil
}

// Parse a response object this is the start of the parsing process from the handler
func (resp Response) parse() interface{} {
	logger.Debug("Building payload for", slog.Any("status", resp.StatusCode), slog.Any("description", resp.Description))

	// Simple case 1: Response has examples defined per content type
	if resp.Examples != nil {
		// We look for an example matching the content type of application/json
		ex := resp.Examples[contentType]
		if ex != nil {
			return ex
		} else {
			logger.Warn("No response example found for content type", slog.Any("content_type", contentType))
		}
	}

	// Complex case: We need to go down the rabbit hole of the schema
	return resp.Schema.parse()
}

// The lowest level of the parser - parses a map of properties looking for examples
// If no example is found, it will create one based on the type
func parseProperties(properties map[string]Properties) interface{} {
	payload := make(map[string]interface{})

	for key, prop := range properties {
		var exampleVal any
		if prop.Example == nil {
			switch prop.Type {
			case "string":
				exampleVal = "string"
			case "integer":
				exampleVal = 0
			case "boolean":
				exampleVal = false
			case "array":
				exampleVal = []string{}
			case "object":
				// Recurse down the rabbit hole of sub-properties
				if prop.Properties != nil {
					exampleVal = parseProperties(prop.Properties)
				} else {
					exampleVal = make(map[string]interface{})
				}
			}
		} else {
			exampleVal = prop.Example
		}

		payload[key] = exampleVal
	}

	return payload
}

// Some helper functions to make the code more readable

func (p PathSpec) isGet() bool {
	return p.Get.Description != "" || p.Get.Responses != nil
}

func (p PathSpec) isPost() bool {
	return p.Post.Responses != nil || p.Post.Description != ""
}

func (p PathSpec) isPut() bool {
	return p.Put.Responses != nil || p.Put.Description != ""
}

func (p PathSpec) isDelete() bool {
	return p.Delete.Responses != nil || p.Delete.Description != ""
}

func (s Schema) isRef() bool {
	return s.Ref != ""
}
