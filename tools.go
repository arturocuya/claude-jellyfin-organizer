package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

var ListDirectoryDefinition = ToolDefinition{
	Name:        "list_directory",
	Description: "List the contents of a directory. Returns file and directory names in the specified path.",
	InputSchema: ListDirectoryInputSchema,
	Function:    ListDirectory,
}

var SearchIMDbDefinition = ToolDefinition{
	Name:        "search_imdb",
	Description: "Search for a title on IMDb. Returns a JSON string with title, id, and description.",
	InputSchema: SearchIMDbInputSchema,
	Function:    SearchIMDb,
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

type ListDirectoryInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a directory in the working directory. Leave empty for current directory."`
}

type SearchIMDbInput struct {
	SearchTerm string `json:"search_term" jsonschema_description:"The search term to look for on IMDb."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()
var ListDirectoryInputSchema = GenerateSchema[ListDirectoryInput]()
var SearchIMDbInputSchema = GenerateSchema[SearchIMDbInput]()

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		panic(err)
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func ListDirectory(input json.RawMessage) (string, error) {
	listDirInput := ListDirectoryInput{}
	err := json.Unmarshal(input, &listDirInput)
	if err != nil {
		return "", err
	}

	dirPath := listDirInput.Path
	if dirPath == "" {
		dirPath = "."
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	result := ""
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			result += name + "/\n"
		} else {
			info, err := entry.Info()
			if err != nil {
				result += name + "\n"
			} else {
				result += fmt.Sprintf("%s (%d bytes)\n", name, info.Size())
			}
		}
	}

	return result, nil
}

func SearchIMDb(input json.RawMessage) (string, error) {
	searchInput := SearchIMDbInput{}
	err := json.Unmarshal(input, &searchInput)
	if err != nil {
		return "", err
	}

	// TODO: Implement IMDb search
	// For now, return empty JSON structure
	return `{"title":"","id":"","description":""}`, nil
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}
