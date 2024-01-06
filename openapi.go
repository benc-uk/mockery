package main

import (
	"encoding/json"
	"io"
	"os"
)

type OpenAPIv2 struct {
	Swagger     string              `json:"swagger"`
	Info        Info                `json:"info"`
	Host        string              `json:"host"`
	BasePath    string              `json:"basePath"`
	Paths       map[string]PathSpec `json:"paths"`
	Definitions map[string]Schema   `json:"definitions"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type PathSpec struct {
	Get        Operation    `json:"get"`
	Post       Operation    `json:"post"`
	Put        Operation    `json:"put"`
	Delete     Operation    `json:"delete"`
	Patch      Operation    `json:"patch"`
	Parameters []Parameters `json:"parameters"`
}

type Operation struct {
	Tags        []string     `json:"tags"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	OperationId string       `json:"operationId"`
	Consumes    []string     `json:"consumes"`
	Produces    []string     `json:"produces"`
	Parameters  []Parameters `json:"parameters"`
	Responses   Responses    `json:"responses"`
}

type Parameters struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

type Responses map[string]Response

type Response struct {
	Description string         `json:"description"`
	Schema      Schema         `json:"schema"`
	Examples    map[string]any `json:"examples"`
}

type Schema struct {
	Example    interface{}           `json:"example"`
	Type       string                `json:"type"`
	Items      Items                 `json:"items"`
	Properties map[string]Properties `json:"properties"`
	Ref        string                `json:"$ref"`
}

type Items struct {
	Type       string                `json:"type"`
	Properties map[string]Properties `json:"properties"`
}

type Properties struct {
	Type    string      `json:"type"`
	Example interface{} `json:"example"`
}

func ParseV2Spec(path string) (OpenAPIv2, error) {
	file, err := os.Open(path)
	if err != nil {
		return OpenAPIv2{}, err
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return OpenAPIv2{}, err
	}

	var openAPIv2 OpenAPIv2
	err = json.Unmarshal(data, &openAPIv2)

	return openAPIv2, err
}

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
