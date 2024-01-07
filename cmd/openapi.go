package main

// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// Mockery - OpenAPI spec structures
// ----------------------------------------------------------------------------

type OpenAPIv2 struct {
	Swagger     string              `json:"swagger" yaml:"swagger"`
	Info        Info                `json:"info" yaml:"info"`
	Host        string              `json:"host" yaml:"host"`
	BasePath    string              `json:"basePath" yaml:"basePath"`
	Paths       map[string]PathSpec `json:"paths" yaml:"paths"`
	Definitions map[string]Schema   `json:"definitions" yaml:"definitions"`
}

type Info struct {
	Title   string `json:"title" yaml:"title"`
	Version string `json:"version" yaml:"version"`
}

type PathSpec struct {
	Get        Operation    `json:"get" yaml:"get"`
	Post       Operation    `json:"post" yaml:"post"`
	Put        Operation    `json:"put" yaml:"put"`
	Delete     Operation    `json:"delete" yaml:"delete"`
	Patch      Operation    `json:"patch" yaml:"patch"`
	Parameters []Parameters `json:"parameters" yaml:"parameters"`
}

type Operation struct {
	Tags        []string     `json:"tags" yaml:"tags"`
	Summary     string       `json:"summary" yaml:"summary"`
	Description string       `json:"description" yaml:"description"`
	OperationID string       `json:"operationId" yaml:"operationId"`
	Consumes    []string     `json:"consumes" yaml:"consumes"`
	Produces    []string     `json:"produces" yaml:"produces"`
	Parameters  []Parameters `json:"parameters" yaml:"parameters"`
	Responses   Responses    `json:"responses" yaml:"responses"`
}

type Parameters struct {
	Name        string `json:"name" yaml:"name"`
	In          string `json:"in" yaml:"in"`
	Description string `json:"description" yaml:"description"`
	Required    bool   `json:"required" yaml:"required"`
	Schema      Schema `json:"schema" yaml:"schema"`
}

type Responses map[string]Response

type Response struct {
	Description string         `json:"description" yaml:"description"`
	Schema      Schema         `json:"schema" yaml:"schema"`
	Examples    map[string]any `json:"examples" yaml:"examples"`
	StatusCode  int            `json:"-" yaml:"-"`
}

type Schema struct {
	Example              interface{}           `json:"example" yaml:"example"`
	Type                 string                `json:"type" yaml:"type"`
	Items                Items                 `json:"items" yaml:"items"`
	Properties           map[string]Properties `json:"properties" yaml:"properties"`
	Ref                  string                `json:"$ref" yaml:"$ref"`
	AdditionalProperties any                   `json:"additionalProperties" yaml:"additionalProperties"`
}

type Items struct {
	Type       string                `json:"type" yaml:"type"`
	Properties map[string]Properties `json:"properties" yaml:"properties"`
	Ref        string                `json:"$ref" yaml:"$ref"`
}

type Properties struct {
	Type       string                `json:"type" yaml:"type"`
	Example    interface{}           `json:"example" yaml:"example"`
	Properties map[string]Properties `json:"properties" yaml:"properties"`
}

func (s Schema) isEmpty() bool {
	return s.Type == "" && s.Ref == "" && s.Properties == nil &&
		s.Items.isEmpty() && s.AdditionalProperties == nil && s.Example == nil
}

func (i Items) isEmpty() bool {
	return i.Type == "" && i.Ref == "" && i.Properties == nil
}
